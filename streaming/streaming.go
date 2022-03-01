// Package streaming contains structure and methods for uploading and downloading files from LocalEGA instance.
package streaming

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/buger/jsonparser"
	"github.com/cheggaaa/pb/v3"
	"github.com/elixir-oslo/crypt4gh/model/headers"
	"github.com/elixir-oslo/lega-commander/conf"
	"github.com/elixir-oslo/lega-commander/files"
	"github.com/elixir-oslo/lega-commander/requests"
	"github.com/elixir-oslo/lega-commander/resuming"
	"github.com/golang-jwt/jwt"
	aurora "github.com/logrusorgru/aurora/v3"
)

// Streamer interface provides methods for uploading and downloading files from LocalEGA instance.
type Streamer interface {
	Upload(path string, resume bool, straight bool) error
	uploadFolder(folder *os.File, resume bool, straight bool) error
	uploadFile(file *os.File, stat os.FileInfo, uploadID *string, offset int64, startChunk int64) error
	Download(fileName string) error
}

type defaultStreamer struct {
	client            requests.Client
	fileManager       files.FileManager
	resumablesManager resuming.ResumablesManager
}

// NewStreamer method constructs Streamer structure.
func NewStreamer(client *requests.Client, fileManager *files.FileManager, resumablesManager *resuming.ResumablesManager) (Streamer, error) {
	streamer := defaultStreamer{}
	if client != nil {
		streamer.client = *client
	} else {
		streamer.client = requests.NewClient(nil)
	}
	if fileManager != nil {
		streamer.fileManager = *fileManager
	} else {
		newFileManager, err := files.NewFileManager(&streamer.client)
		if err != nil {
			return nil, err
		}
		streamer.fileManager = newFileManager
	}
	if resumablesManager != nil {
		streamer.resumablesManager = *resumablesManager
	} else {
		newResumablesManager, err := resuming.NewResumablesManager(&streamer.client)
		if err != nil {
			return nil, err
		}
		streamer.resumablesManager = newResumablesManager
	}
	return streamer, nil
}

// Upload method uploads file or folder to LocalEGA.
func (s defaultStreamer) Upload(path string, resume bool) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return err
	}
	if stat.IsDir() {
		return s.uploadFolder(file, resume)
	}
	if resume {
		fileName := filepath.Base(file.Name())
		resumablesList, err := s.resumablesManager.ListResumables()
		if err != nil {
			return err
		}
		for _, resumable := range *resumablesList {
			if resumable.Name == fileName {
				return s.uploadFile(file, stat, &resumable.ID, resumable.Size, resumable.Chunk)
			}
		}
		return nil
	}
	return s.uploadFile(file, stat, nil, 0, 1)
}

func (s defaultStreamer) uploadFolder(folder *os.File, resume bool) error {
	readdir, err := folder.Readdir(-1)
	if err != nil {
		return err
	}
	for _, file := range readdir {
		abs, err := filepath.Abs(filepath.Join(folder.Name(), file.Name()))
		if err != nil {
			return err
		}
		err = s.Upload(abs, resume)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s defaultStreamer) uploadFile(file *os.File, stat os.FileInfo, uploadID *string, offset, startChunk int64) error {
	fileName := filepath.Base(file.Name())
	filesList, err := s.fileManager.ListFiles(true)
	if err != nil {
		return err
	}
	for _, uploadedFile := range *filesList {
		if fileName == filepath.Base(uploadedFile.FileName) {
			return errors.New("File " + file.Name() + " is already uploaded. Please, remove it from the Inbox first: lega-commander files -d " + filepath.Base(uploadedFile.FileName))
		}
	}
	if err = isCrypt4GHFile(file); err != nil {
		return err
	}
	totalSize := stat.Size()
	fmt.Println(aurora.Blue("Uploading file: " + file.Name() + " (" + strconv.FormatInt(totalSize, 10) + " bytes)"))
	bar := pb.StartNew(100)
	bar.SetTotal(totalSize)
	bar.SetCurrent(offset)
	bar.Start()
	configuration := conf.NewConfiguration()
	_, err = file.Seek(offset, 0)
	if err != nil {
		return err
	}
	buffer := make([]byte, configuration.GetChunkSize()*1024*1024)
	for i := startChunk; true; i++ {
		read, err := file.Read(buffer)
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		chunk := buffer[:read]
		sum := md5.Sum(chunk)
		params := map[string]string{
			"chunk": strconv.FormatInt(i, 10),
			"md5":   hex.EncodeToString(sum[:16])}
		if i != 1 {
			params["uploadId"] = *uploadID
		}
		response, err := s.client.DoRequest(http.MethodPatch,
			configuration.GetLocalEGAInstanceURL()+"/stream/"+url.QueryEscape(fileName),
			bytes.NewReader(chunk),
			map[string]string{"Proxy-Authorization": "Bearer " + configuration.GetElixirAAIToken()},
			params,
			configuration.GetCentralEGAUsername(),
			configuration.GetCentralEGAPassword())
		if err != nil {
			return err
		}
		if response.StatusCode != 200 {
			return errors.New(response.Status)
		}
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return err
		}
		err = response.Body.Close()
		if err != nil {
			return err
		}
		if uploadID == nil {
			uploadID = new(string)
		}
		*uploadID, err = jsonparser.GetString(body, "id")
		if err != nil {
			return err
		}
		bar.SetCurrent(int64(read)*(i-startChunk+1) + offset)
	}
	bar.SetCurrent(totalSize)
	hashFunction := sha256.New()
	_, err = io.Copy(hashFunction, file)
	if err != nil {
		return err
	}
	checksum := hex.EncodeToString(hashFunction.Sum(nil))
	fmt.Println("assembling different parts of file together in order to make it! Duration varies based on filesize.")
	response, err := s.client.DoRequest(http.MethodPatch,
		configuration.GetLocalEGAInstanceURL()+"/stream/"+url.QueryEscape(fileName),
		nil,
		map[string]string{"Proxy-Authorization": "Bearer " + configuration.GetElixirAAIToken()},
		map[string]string{"uploadId": *uploadID,
			"chunk":    "end",
			"fileSize": strconv.FormatInt(totalSize, 10),
			"sha256":   checksum},
		configuration.GetCentralEGAUsername(),
		configuration.GetCentralEGAPassword())
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return errors.New(response.Status)
	}
	err = response.Body.Close()
	if err != nil {
		return err
	}
	bar.Finish()
	return nil
}

func isCrypt4GHFile(file *os.File) error {
	_, err := headers.ReadHeader(file)
	if err != nil {
		return errors.New(file.Name() + ": " + err.Error())
	}
	err = file.Close()
	if err != nil {
		return err
	}
	reopenedFile, err := os.Open(file.Name())
	if err != nil {
		return err
	}
	*file = *reopenedFile
	return err
}

// Download method downloads file from LocalEGA.
func (s defaultStreamer) Download(fileName string) error {
	if fileExists(fileName) {
		return errors.New("File " + fileName + " exists locally, aborting.")
	}
	filesList, err := s.fileManager.ListFiles(false)
	if err != nil {
		return err
	}
	found := false
	fileSize := int64(0)
	for _, exportedFile := range *filesList {
		if fileName == filepath.Base(exportedFile.FileName) {
			found = true
			fileSize = exportedFile.Size
			break
		}
	}
	if !found {
		return errors.New("File " + fileName + " not found in the outbox.")
	}
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Println(aurora.Blue("Downloading file: " + file.Name() + " (" + strconv.FormatInt(fileSize, 10) + " bytes)"))
	bar := pb.Start64(fileSize)
	configuration := conf.NewConfiguration()
	response, err := s.client.DoRequest(http.MethodGet,
		configuration.GetLocalEGAInstanceURL()+"/stream/"+url.QueryEscape(fileName),
		nil,
		map[string]string{"Proxy-Authorization": "Bearer " + configuration.GetElixirAAIToken()},
		map[string]string{"fileName": fileName},
		"",
		"")
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return errors.New(response.Status)
	}
	barReader := bar.NewProxyReader(response.Body)
	defer barReader.Close()
	_, err = io.Copy(file, barReader)
	return err
}

func fileExists(fileName string) bool {
	info, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
