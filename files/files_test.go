package files

import (
	"github.com/elixir-oslo/lega-commander/requests"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"
)

type mockClient struct {
}

func (mockClient) DoRequest(method, url string, _ io.Reader, _, params map[string]string, _, _ string) (*http.Response, error) {
	if strings.HasSuffix(url, "/files") {
		if method == http.MethodGet {
			var body io.ReadCloser
			if params["inbox"] == "" || params["inbox"] == "true" {
				body = ioutil.NopCloser(strings.NewReader(`{"files": [{"fileName": "test.enc", "size": 100, "modifiedDate": "2010"}]}`))
			} else {
				body = ioutil.NopCloser(strings.NewReader(`{"files": [{"fileName": "test2.enc", "size": 100, "modifiedDate": "2010"}]}`))
			}
			response := http.Response{StatusCode: 200, Body: body}
			return &response, nil
		} else if method == http.MethodDelete {
			if params["inbox"] == "false" {
				response := http.Response{StatusCode: 500}
				return &response, nil
			}
			if params["fileName"] == "test.enc" {
				response := http.Response{StatusCode: 200}
				return &response, nil
			}
			response := http.Response{StatusCode: 500}
			return &response, nil
		}
	}
	return nil, nil
}

func TestListFilesInbox(t *testing.T) {
	_ = os.Setenv("CENTRAL_EGA_USERNAME", "user")
	_ = os.Setenv("CENTRAL_EGA_PASSWORD", "pass")
	_ = os.Setenv("LOCAL_EGA_INSTANCE_URL", "http://localhost/")
	_ = os.Setenv("ELIXIR_AAI_TOKEN", "token")
	var client requests.Client = mockClient{}
	fileManager, err := NewFileManager(&client)
	if err != nil {
		t.Error(err)
	}
	fileList, err := fileManager.ListFiles(true)
	if err != nil {
		t.Error(err)
	}
	if fileList == nil || len(*fileList) != 1 {
		t.Error()
	}
	file := (*fileList)[0]
	if file.FileName != "test.enc" || file.Size != 100 || file.ModifiedDate != "2010" {
		t.Error()
	}
}

func TestListFilesOutbox(t *testing.T) {
	_ = os.Setenv("CENTRAL_EGA_USERNAME", "user")
	_ = os.Setenv("CENTRAL_EGA_PASSWORD", "pass")
	_ = os.Setenv("LOCAL_EGA_INSTANCE_URL", "http://localhost/")
	_ = os.Setenv("ELIXIR_AAI_TOKEN", "token")
	var client requests.Client = mockClient{}
	fileManager, err := NewFileManager(&client)
	if err != nil {
		t.Error(err)
	}
	fileList, err := fileManager.ListFiles(false)
	if err != nil {
		t.Error(err)
	}
	if fileList == nil || len(*fileList) != 1 {
		t.Error()
	}
	file := (*fileList)[0]
	if file.FileName != "test2.enc" || file.Size != 100 || file.ModifiedDate != "2010" {
		t.Error()
	}
}

func TestDeleteInboxFile200(t *testing.T) {
	_ = os.Setenv("CENTRAL_EGA_USERNAME", "user")
	_ = os.Setenv("CENTRAL_EGA_PASSWORD", "pass")
	_ = os.Setenv("LOCAL_EGA_INSTANCE_URL", "http://localhost/")
	_ = os.Setenv("ELIXIR_AAI_TOKEN", "token")
	var client requests.Client = mockClient{}
	fileManager, err := NewFileManager(&client)
	if err != nil {
		t.Error(err)
	}
	err = fileManager.DeleteFile("test.enc")
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteInboxFile500(t *testing.T) {
	_ = os.Setenv("CENTRAL_EGA_USERNAME", "user")
	_ = os.Setenv("CENTRAL_EGA_PASSWORD", "pass")
	_ = os.Setenv("LOCAL_EGA_INSTANCE_URL", "http://localhost/")
	_ = os.Setenv("ELIXIR_AAI_TOKEN", "token")
	var client requests.Client = mockClient{}
	fileManager, err := NewFileManager(&client)
	if err != nil {
		t.Error(err)
	}
	err = fileManager.DeleteFile("12")
	if err == nil {
		t.Error(err)
	}
}
