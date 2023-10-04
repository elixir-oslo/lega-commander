// Package files contains structures and methods for listing or deleting uploaded files.
package files

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/elixir-oslo/lega-commander/conf"
	"github.com/elixir-oslo/lega-commander/requests"
)

// File structure represents uploaded File.
type File struct {
	FileName     string
	Size         int64
	ModifiedDate string
}

// FileManager interface provides method for managing uploaded files.
type FileManager interface {
	ListFiles(inbox bool) (*[]File, error)
	DeleteFile(fileName string) error
}

type defaultFileManager struct {
	client requests.Client
}

// Represents an error message to handle a missing or empty folder
type FolderNotFoundError struct {
    Msg string
}

// Returns the error message in FolderNotFoundError.
func (e *FolderNotFoundError) Error() string {
    return e.Msg
}

// NewFileManager constructs FileManager using requests.Client.
func NewFileManager(client *requests.Client) (FileManager, error) {
	fileManager := defaultFileManager{}
	if client != nil {
		fileManager.client = *client
	} else {
		fileManager.client = requests.NewClient(nil)
	}
	return fileManager, nil
}

// ListFiles method lists uploaded files.
func (rm defaultFileManager) ListFiles(inbox bool) (*[]File, error) {
	configuration := conf.NewConfiguration()
	username, password := "", ""
	if inbox {
		username = configuration.GetCentralEGAUsername()
		password = configuration.GetCentralEGAPassword()
	}
	response, err := rm.client.DoRequest(http.MethodGet,
		configuration.GetLocalEGAInstanceURL()+"/files",
		nil,
		map[string]string{"Proxy-Authorization": "Bearer " + configuration.GetElixirAAIToken()},
		map[string]string{"inbox": strconv.FormatBool(inbox)},
		username,
		password)
	if err != nil {
		return nil, err
	}
	if response.StatusCode == 403 {
    		body, err := ioutil.ReadAll(response.Body)
    		if err != nil {
    			return nil, errors.New("Failed to read the server's response")
    		}
    		// Check if the response body contains the specific error message indicating
    		// that the folder is empty or doesn't exist yet.
    		if strings.Contains(string(body), `"tsdFiles" is null`) {
    			return nil, &FolderNotFoundError{}
    		}
    		// If it's not an empty folder, it's a genuine authentication error.
    		return nil, errors.New("Authentication error")
    	} else if response.StatusCode != 200 {
    	return nil, errors.New(response.Status)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	err = response.Body.Close()
	if err != nil {
		return nil, err
	}
	files := make([]File, 0)
	_, _ = jsonparser.ArrayEach(body,
		func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			fileName, _ := jsonparser.GetString(value, "fileName")
			size, _ := jsonparser.GetInt(value, "size")
			modifiedDate, _ := jsonparser.GetString(value, "modifiedDate")
			file := File{fileName, size, modifiedDate}
			files = append(files, file)
		},
		"files")
	return &files, nil
}

// DeleteFiles method deletes uploaded file by its name.
func (rm defaultFileManager) DeleteFile(fileName string) error {
	configuration := conf.NewConfiguration()
	response, err := rm.client.DoRequest(http.MethodDelete,
		configuration.GetLocalEGAInstanceURL()+"/files",
		nil,
		map[string]string{"Proxy-Authorization": "Bearer " + configuration.GetElixirAAIToken()},
		map[string]string{"fileName": fileName},
		configuration.GetCentralEGAUsername(),
		configuration.GetCentralEGAPassword())
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return errors.New(response.Status)
	}
	return nil
}
