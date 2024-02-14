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
	"github.com/elixir-oslo/lega-commander/conf"
	"github.com/elixir-oslo/lega-commander/files"
	"github.com/elixir-oslo/lega-commander/requests"
	"github.com/elixir-oslo/lega-commander/resuming"
	"github.com/golang-jwt/jwt"
	aurora "github.com/logrusorgru/aurora/v3"
	"github.com/neicnordic/crypt4gh/model/headers"
)

// Streamer interface provides methods for uploading and downloading files from LocalEGA instance.
type Streamer interface {
	Upload(path string, resume, straight bool) error
	uploadFolder(folder *os.File, resume bool, straight bool) error
	uploadFile(file *os.File, stat os.FileInfo, uploadID *string, offset int64, startChunk int64) error
	Download(fileName string) error
}

type defaultStreamer struct {
	client            requests.Client
	fileManager       files.FileManager
	resumablesManager resuming.ResumablesManager
	tsd_token         string
	claims            jwt.MapClaims
}
type ResponseJson struct {
	// defining token response that comes from tsd proxy
	StatusCode int    `json:"statusCode"`
	StatusText string `json:"statusText"`
	Token      string `json:"token"`
}

// NewStreamer method constructs Streamer structure.
func NewStreamer(client *requests.Client, fileManager *files.FileManager, resumablesManager *resuming.ResumablesManager, straight bool) (Streamer, error) {
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
	configuration := conf.NewConfiguration()
	var err error
	if straight {
		streamer.tsd_token, streamer.claims, err = streamer.getTSDtoken(configuration)
	}
	if err != nil {
		return nil, err
	}
	return streamer, nil
}

// Upload method uploads file or folder to LocalEGA.
func (s defaultStreamer) Upload(path string, resume, straight bool) error {
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
		return s.uploadFolder(file, resume, straight)
	}
	if resume {
		fileName := filepath.Base(file.Name())
		resumablesList, err := s.resumablesManager.ListResumables()
		if err != nil {
			return err
		}
		for _, resumable := range *resumablesList {
			if resumable.Name == fileName {
				if !straight {
					return s.uploadFile(file, stat, &resumable.ID, resumable.Size, resumable.Chunk)
				} else {
					return s.uploadFileWithoutProxy(file, stat, &resumable.ID, resumable.Size, resumable.Chunk)
				}

			}
		}
		return nil
	}
	if !straight {
		return s.uploadFile(file, stat, nil, 0, 1)
	} else {
		return s.uploadFileWithoutProxy(file, stat, nil, 0, 1)
	}
}

func (s defaultStreamer) uploadFolder(folder *os.File, resume, straight bool) error {
	readdir, err := folder.Readdir(-1)
	if err != nil {
		return err
	}
	for _, file := range readdir {
		abs, err := filepath.Abs(filepath.Join(folder.Name(), file.Name()))
		if err != nil {
			return err
		}
		err = s.Upload(abs, resume, straight)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s defaultStreamer) uploadFile(file *os.File, stat os.FileInfo, uploadID *string, offset, startChunk int64) error {
	fileName := filepath.Base(file.Name())

	// List user's files already in inbox to avoid accidental overwrites
	filesList, err := s.fileManager.ListFiles(true)
	if err != nil {
		fmt.Println("Could not read previous uploaded files, this is ok if it's your first upload")
		//		return err
	} else {
		for _, uploadedFile := range *filesList {
			if fileName == filepath.Base(uploadedFile.FileName) {
				return errors.New("File " + file.Name() + " is already uploaded. Please, remove it from the Inbox first: lega-commander files -d " + filepath.Base(uploadedFile.FileName))
			}
		}
	}

	// Make sure the file to be uploaded is a crypt4gh encrypted file
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
	hashFunction := sha256.New()
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
		hashFunction.Write(chunk)
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
	checksum := hex.EncodeToString(hashFunction.Sum(nil))
	fmt.Println("Assembling the uploaded parts of the file together in order to build it! Duration varies based on filesize.")
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

func extractClaims(response *http.Response) (string, jwt.MapClaims, error) {
	if response.StatusCode != 200 {
		return "", nil, errors.New(response.Status)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", nil, err
	}

	var respjson ResponseJson
	err = json.Unmarshal(body, &respjson)
	if err != nil {
		return "", nil, err
	}
	err = response.Body.Close()
	if err != nil {
		return "", nil, err
	}

	claims := jwt.MapClaims{}
	jwt.ParseWithClaims(respjson.Token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(""), nil
	})
	return respjson.Token, claims, nil
}

func (s defaultStreamer) getTSDtoken(c conf.Configuration) (string, jwt.MapClaims, error) {
	fmt.Println("asking for tsd connection details from proxy service...")
	// var response *http.Response
	// var err error
	response, err := s.client.DoRequest(http.MethodGet,
		c.GetLocalEGAInstanceURL()+"/gettoken",
		nil,
		map[string]string{"Proxy-Authorization": "Bearer " + c.GetElixirAAIToken()},
		nil,
		c.GetCentralEGAUsername(),
		c.GetCentralEGAPassword())
	if err != nil {
		return "", nil, err
	}
	return extractClaims(response)
}

func (s *defaultStreamer) uploadFileWithoutProxy(file *os.File, stat os.FileInfo, uploadID *string, offset, startChunk int64) error {
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
	configuration := conf.NewConfiguration()
	streamurl := configuration.ConcatenateURLPartsToString(
		[]string{
			configuration.GetTSDURL(), s.claims["user"].(string), "files", url.QueryEscape(fileName),
		},
	)
	if err = isCrypt4GHFile(file); err != nil {
		return err
	}
	totalSize := stat.Size()
	fmt.Println(aurora.Blue("Uploading file: " + file.Name() + " (" + strconv.FormatInt(totalSize, 10) + " bytes)"))
	bar := pb.StartNew(100)
	bar.SetTotal(totalSize)
	bar.SetCurrent(offset)
	bar.Start()

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
		if err != nil {
			return err
		}
		var response *http.Response
		if i != 1 {
			response, err = s.client.DoRequest(http.MethodPatch,
				streamurl,
				bytes.NewReader(chunk),
				map[string]string{"Authorization": "Bearer " + s.tsd_token},
				map[string]string{"id": *uploadID, "chunk": strconv.FormatInt(i, 10)},
				"",
				"")
		} else {
			response, err = s.client.DoRequest(http.MethodPatch,
				streamurl,
				bytes.NewReader(chunk),
				map[string]string{"Authorization": "Bearer " + s.tsd_token},
				map[string]string{"chunk": "1"},
				"",
				"")
		}
		if err != nil {
			return err
		}
		if !(response.StatusCode == 200 || response.StatusCode == 201) {
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
	fmt.Println("assembling different parts of file together in order to make it! Duration varies based on filesize.")
	response, err := s.client.DoRequest(http.MethodPatch,
		streamurl,
		nil,
		map[string]string{"Authorization": "Bearer " + s.tsd_token},
		map[string]string{"id": *uploadID, "chunk": "end"},
		"",
		"")
	if err != nil {
		return err
	}
	if !(response.StatusCode == 200 || response.StatusCode == 201) {
		return errors.New(response.Status)
	}
	err = response.Body.Close()
	if err != nil {
		return err
	}
	bar.Finish()
	return nil
}
