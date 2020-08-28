// Package resuming contains structures and methods for listing or resuming uploads.
package resuming

import (
	"errors"
	"github.com/buger/jsonparser"
	"github.com/uio-bmi/lega-uploader/conf"
	"github.com/uio-bmi/lega-uploader/requests"
	"io/ioutil"
	"net/http"
)

// Resumable structure represents resumable upload.
type Resumable struct {
	ID    string
	Name  string
	Size  int64
	Chunk int64
}

// ResumablesManager interface provides method for managing resumable uploads.
type ResumablesManager interface {
	ListResumables() (*[]Resumable, error)
	DeleteResumable(uploadID string) error
}

type defaultResumablesManager struct {
	client requests.Client
}

// NewResumablesManager constructs ResumablesManager using requests.Client.
func NewResumablesManager(client *requests.Client) (ResumablesManager, error) {
	resumablesManager := defaultResumablesManager{}
	if client != nil {
		resumablesManager.client = *client
	} else {
		resumablesManager.client = requests.NewClient(nil)
	}
	return resumablesManager, nil
}

// ListResumables method lists resumable uploads.
func (rm defaultResumablesManager) ListResumables() (*[]Resumable, error) {
	configuration := conf.NewConfiguration()
	response, err := rm.client.DoRequest(http.MethodGet,
		configuration.GetLocalEGAInstanceURL()+"/resumables",
		nil,
		map[string]string{"Proxy-Authorization": "Bearer " + configuration.GetElixirAAIToken()},
		nil,
		configuration.GetCentralEGAUsername(),
		configuration.GetCentralEGAPassword())
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
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
	resumables := make([]Resumable, 0)
	_, err = jsonparser.ArrayEach(body,
		func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
			id, _ := jsonparser.GetString(value, "id")
			fileName, _ := jsonparser.GetString(value, "fileName")
			size, _ := jsonparser.GetInt(value, "nextOffset")
			chunk, _ := jsonparser.GetInt(value, "maxChunk")
			chunk++
			resumable := Resumable{id, fileName, size, chunk}
			resumables = append(resumables, resumable)
		},
		"resumables")
	if err != nil {
		return nil, err
	}
	return &resumables, nil
}

// DeleteResumable method deletes resumable upload by its ID.
func (rm defaultResumablesManager) DeleteResumable(uploadID string) error {
	configuration := conf.NewConfiguration()
	response, err := rm.client.DoRequest(http.MethodDelete,
		configuration.GetLocalEGAInstanceURL()+"/resumables",
		nil,
		map[string]string{"Proxy-Authorization": "Bearer " + configuration.GetElixirAAIToken()},
		map[string]string{"uploadId": uploadID},
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
