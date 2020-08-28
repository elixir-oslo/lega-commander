package resuming

import (
	"github.com/uio-bmi/lega-uploader/requests"
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
	if strings.HasSuffix(url, "/resumables") {
		if method == http.MethodGet {
			body := ioutil.NopCloser(strings.NewReader(`{"resumables": [{"id": "1", "fileName": "test.enc", "nextOffset": 100, "maxChunk": 10}]}`))
			response := http.Response{StatusCode: 200, Body: body}
			return &response, nil
		} else if method == http.MethodDelete {
			if params["uploadId"] == "123" {
				response := http.Response{StatusCode: 200}
				return &response, nil
			}
			response := http.Response{StatusCode: 500}
			return &response, nil
		}
	}
	return nil, nil
}

func TestListResumables(t *testing.T) {
	_ = os.Setenv("CENTRAL_EGA_USERNAME", "user")
	_ = os.Setenv("CENTRAL_EGA_PASSWORD", "pass")
	_ = os.Setenv("LOCAL_EGA_INSTANCE_URL", "http://localhost/")
	_ = os.Setenv("ELIXIR_AAI_TOKEN", "token")
	var client requests.Client = mockClient{}
	resumablesManager, err := NewResumablesManager(&client)
	if err != nil {
		t.Error(err)
	}
	resumables, err := resumablesManager.ListResumables()
	if err != nil {
		t.Error(err)
	}
	if resumables == nil || len(*resumables) != 1 {
		t.Error()
	}
	resumable := (*resumables)[0]
	if resumable.ID != "1" || resumable.Name != "test.enc" || resumable.Size != 100 || resumable.Chunk != 11 {
		t.Error()
	}
}

func TestDeleteResumable200(t *testing.T) {
	_ = os.Setenv("CENTRAL_EGA_USERNAME", "user")
	_ = os.Setenv("CENTRAL_EGA_PASSWORD", "pass")
	_ = os.Setenv("LOCAL_EGA_INSTANCE_URL", "http://localhost/")
	_ = os.Setenv("ELIXIR_AAI_TOKEN", "token")
	var client requests.Client = mockClient{}
	resumablesManager, err := NewResumablesManager(&client)
	if err != nil {
		t.Error(err)
	}
	err = resumablesManager.DeleteResumable("123")
	if err != nil {
		t.Error(err)
	}
}

func TestDeleteResumable500(t *testing.T) {
	_ = os.Setenv("CENTRAL_EGA_USERNAME", "user")
	_ = os.Setenv("CENTRAL_EGA_PASSWORD", "pass")
	_ = os.Setenv("LOCAL_EGA_INSTANCE_URL", "http://localhost/")
	_ = os.Setenv("ELIXIR_AAI_TOKEN", "token")
	var client requests.Client = mockClient{}
	resumablesManager, err := NewResumablesManager(&client)
	if err != nil {
		t.Error(err)
	}
	err = resumablesManager.DeleteResumable("12")
	if err == nil {
		t.Error(err)
	}
}
