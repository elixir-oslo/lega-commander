// Package uploading contains structure and methods for uploading files towards LocalEGA instance.
package uploading

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/cheggaaa/pb/v3"
	"github.com/elixir-oslo/crypt4gh/model/headers"
	"github.com/logrusorgru/aurora"
	"github.com/uio-bmi/lega-uploader/conf"
	"github.com/uio-bmi/lega-uploader/files"
	"github.com/uio-bmi/lega-uploader/requests"
	"github.com/uio-bmi/lega-uploader/resuming"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
)

// Uploader interface provides methods for uploading files towards LocalEGA instance.
type Uploader interface {
	Upload(path string, resume bool) error
	uploadFolder(folder *os.File, resume bool) error
	uploadFile(file *os.File, stat os.FileInfo, uploadID *string, offset int64, startChunk int64) error
}

type defaultUploader struct {
	client            requests.Client
	fileManager       files.FileManager
	resumablesManager resuming.ResumablesManager
}

// NewUploader method constructs Uploader structure.
func NewUploader(client *requests.Client, fileManager *files.FileManager, resumablesManager *resuming.ResumablesManager) (Uploader, error) {
	uploader := defaultUploader{}
	if client != nil {
		uploader.client = *client
	} else {
		uploader.client = requests.NewClient(nil)
	}
	if fileManager != nil {
		uploader.fileManager = *fileManager
	} else {
		newFileManager, err := files.NewFileManager(&uploader.client)
		if err != nil {
			return nil, err
		}
		uploader.fileManager = newFileManager
	}
	if resumablesManager != nil {
		uploader.resumablesManager = *resumablesManager
	} else {
		newResumablesManager, err := resuming.NewResumablesManager(&uploader.client)
		if err != nil {
			return nil, err
		}
		uploader.resumablesManager = newResumablesManager
	}
	return uploader, nil
}

// Upload method uploads file or folder to LocalEGA.
func (u defaultUploader) Upload(path string, resume bool) error {
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
		return u.uploadFolder(file, resume)
	}
	if resume {
		fileName := filepath.Base(file.Name())
		resumablesList, err := u.resumablesManager.ListResumables()
		if err != nil {
			return err
		}
		for _, resumable := range *resumablesList {
			if resumable.Name == fileName {
				return u.uploadFile(file, stat, &resumable.ID, resumable.Size, resumable.Chunk)
			}
		}
		return nil
	}
	return u.uploadFile(file, stat, nil, 0, 1)
}

func (u defaultUploader) uploadFolder(folder *os.File, resume bool) error {
	readdir, err := folder.Readdir(-1)
	if err != nil {
		return err
	}
	for _, file := range readdir {
		abs, err := filepath.Abs(filepath.Join(folder.Name(), file.Name()))
		if err != nil {
			return err
		}
		err = u.Upload(abs, resume)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u defaultUploader) uploadFile(file *os.File, stat os.FileInfo, uploadID *string, offset, startChunk int64) error {
	fileName := filepath.Base(file.Name())
	filesList, err := u.fileManager.ListFiles()
	if err != nil {
		return err
	}
	for _, uploadedFile := range *filesList {
		if fileName == filepath.Base(uploadedFile.FileName) {
			return errors.New("File " + file.Name() + " is already uploaded. Please, remove it from the Inbox first: lega-uploader files -d " + filepath.Base(uploadedFile.FileName))
		}
	}
	if err = isCrypt4GHFile(file); err != nil {
		return err
	}
	totalSize := stat.Size()
	fmt.Println(aurora.Blue("Uploading file: " + file.Name() + " (" + strconv.FormatInt(totalSize, 10) + " bytes)"))
	bar := pb.StartNew(100)
	bar.SetCurrent(offset * 100 / totalSize)
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
		response, err := u.client.DoRequest(http.MethodPatch,
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
		bar.Add64(int64(read) * 100 / totalSize)
	}
	bar.SetCurrent(100)
	hashFunction := sha256.New()
	_, err = io.Copy(hashFunction, file)
	if err != nil {
		return err
	}
	checksum := hex.EncodeToString(hashFunction.Sum(nil))
	response, err := u.client.DoRequest(http.MethodPatch,
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
