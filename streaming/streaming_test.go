package streaming

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/chzyer/test"
	"github.com/elixir-oslo/lega-commander/files"
	"github.com/elixir-oslo/lega-commander/requests"
	"github.com/elixir-oslo/lega-commander/resuming"
	aurora "github.com/logrusorgru/aurora/v3"
)

var uploader Streamer
var dir string
var file *os.File
var existingFile *os.File

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	var err error
	_ = os.Setenv("CENTRAL_EGA_USERNAME", "user")
	_ = os.Setenv("CENTRAL_EGA_PASSWORD", "pass")
	_ = os.Setenv("LOCAL_EGA_INSTANCE_URL", "http://localhost/")
	_ = os.Setenv("ELIXIR_AAI_TOKEN", "token")
	var client requests.Client = mockClient{}
	filesManager, err := files.NewFileManager(&client)
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	resumablesManager, err := resuming.NewResumablesManager(&client)
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	uploader, err = NewStreamer(&client, &filesManager, &resumablesManager, false)
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	dir = "../test/files"
	file, err = os.Open("../test/files/sample.txt.enc")
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
	existingFile, err = os.Open("../test/test.enc")
	if err != nil {
		log.Fatal(aurora.Red(err))
	}
}

type mockClient struct {
}

func (mockClient) DoRequest(method, url string, _ io.Reader, headers, params map[string]string, _, _ string) (*http.Response, error) {
	var response http.Response
	if !strings.HasPrefix(headers["Proxy-Authorization"], "Bearer ") {
		body := ioutil.NopCloser(strings.NewReader(""))
		response = http.Response{StatusCode: 401, Body: body}
		return &response, nil
	}
	if strings.HasSuffix(url, "/files") {
		var body io.ReadCloser
		if params["inbox"] == "" || params["inbox"] == "true" {
			body = ioutil.NopCloser(strings.NewReader(`{"files": [{"fileName": "test.enc", "size": 100, "modifiedDate": "2010"}]}`))
		} else {
			body = ioutil.NopCloser(strings.NewReader(`{"files": [{"fileName": "test2.enc", "size": 100, "modifiedDate": "2010"}]}`))
		}
		response := http.Response{StatusCode: 200, Body: body}
		return &response, nil
	}
	if strings.Contains(url, "/stream") {
		if method == http.MethodPatch {
			var id string
			var ok bool
			id, ok = params["id"]
			if !ok {
				id = "123"
			}
			if id != "123" {
				body := ioutil.NopCloser(strings.NewReader(""))
				response = http.Response{StatusCode: 500, Body: body}
				return &response, nil
			}
			chunk := params["chunk"]
			if chunk == "1" {
				checksum := params["md5"]
				if checksum != "da385d93ae510bc91c9c8af7e670ac6f" {
					body := ioutil.NopCloser(strings.NewReader(""))
					response = http.Response{StatusCode: 500, Body: body}
					return &response, nil
				}
			} else if chunk == "end" {
				checksum := params["sha256"]
				fileSize := params["fileSize"]
				if checksum != "91e93335245604993d0c2599fa22efc2b607301bf1a9b2426a6489adf6bf5af6" || fileSize != "65688" {
					body := ioutil.NopCloser(strings.NewReader(""))
					response = http.Response{StatusCode: 500, Body: body}
					return &response, nil
				}
			}
			body := ioutil.NopCloser(strings.NewReader(fmt.Sprintf(`{"id":"%s"}`, id)))
			response = http.Response{StatusCode: 200, Body: body}
			return &response, nil
		}
		if method == http.MethodGet {
			body := ioutil.NopCloser(strings.NewReader("test"))
			response = http.Response{StatusCode: 200, Body: body}
			return &response, nil
		}
	}
	return nil, nil
}

func TestUploadedFileExists(t *testing.T) {
	err := uploader.Upload(existingFile.Name(), false, false)
	if err == nil {
		t.Error()
	}
}

func TestUploadFile(t *testing.T) {
	err := uploader.Upload(file.Name(), false, false)
	if err != nil {
		t.Error(err)
	}
}

func TestUploadFolder(t *testing.T) {
	err := uploader.Upload(dir, false, false)
	if err == nil || !strings.HasSuffix(err.Error(), "not a Crypt4GH file") {
		t.Error(err)
	}
}

func TestDownloadFileRemoteDoesntExist(t *testing.T) {
	err := uploader.Download("notfoundfile.enc")
	if err == nil || !strings.HasSuffix(err.Error(), "not found in the outbox.") {
		t.Error(err)
	}
}

func TestDownloadFileRemoteExists(t *testing.T) {
	err := uploader.Download("test2.enc")
	if err != nil {
		t.Error(err)
	}
	file, err := os.Open("test2.enc")
	if err != nil {
		t.Error(err)
	}
	test.ReadString(file, "test")
	err = file.Close()
	if err != nil {
		t.Error(err)
	}
	err = os.Remove("test2.enc")
	if err != nil {
		t.Error(err)
	}
}

func TestDownloadFileLocalExists(t *testing.T) {
	_, err := os.Create("test2.enc")
	if err != nil {
		t.Error(err)
	}
	err = uploader.Download("test2.enc")
	if err == nil || !strings.HasSuffix(err.Error(), "exists locally, aborting.") {
		t.Error(err)
	}
	err = os.Remove("test2.enc")
	if err != nil {
		t.Error(err)
	}
}

func teardown() {
	_ = os.Unsetenv("CENTRAL_EGA_USERNAME")
	_ = os.Unsetenv("CENTRAL_EGA_PASSWORD")
	_ = os.Unsetenv("LOCAL_EGA_INSTANCE_URL")
	_ = os.Unsetenv("ELIXIR_AAI_TOKEN")
}
