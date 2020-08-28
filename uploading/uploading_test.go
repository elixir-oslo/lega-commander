package uploading

import (
	"fmt"
	"github.com/logrusorgru/aurora"
	"github.com/uio-bmi/lega-uploader/files"
	"github.com/uio-bmi/lega-uploader/requests"
	"github.com/uio-bmi/lega-uploader/resuming"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
)

var uploader Uploader
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
	uploader, err = NewUploader(&client, &filesManager, &resumablesManager)
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

func (mockClient) DoRequest(_, url string, _ io.Reader, headers, params map[string]string, username, password string) (*http.Response, error) {
	var response http.Response
	if username != "user" || password != "pass" || !strings.HasPrefix(headers["Proxy-Authorization"], "Bearer ") {
		body := ioutil.NopCloser(strings.NewReader(""))
		response = http.Response{StatusCode: 401, Body: body}
		return &response, nil
	}
	if strings.HasSuffix(url, "/files") {
		body := ioutil.NopCloser(strings.NewReader(`{"files": [{"fileName": "test.enc", "size": 100, "modifiedDate": "2010"}]}`))
		response := http.Response{StatusCode: 200, Body: body}
		return &response, nil
	}
	if strings.Contains(url, "/stream") {
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
			if checksum != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" || fileSize != "65688" {
				body := ioutil.NopCloser(strings.NewReader(""))
				response = http.Response{StatusCode: 500, Body: body}
				return &response, nil
			}
		}
		body := ioutil.NopCloser(strings.NewReader(fmt.Sprintf(`{"id":"%s"}`, id)))
		response = http.Response{StatusCode: 200, Body: body}
		return &response, nil
	}
	return nil, nil
}

func TestUploadedFileExists(t *testing.T) {
	err := uploader.Upload(existingFile.Name(), false)
	if err == nil {
		t.Error()
	}
}

func TestUploadFile(t *testing.T) {
	err := uploader.Upload(file.Name(), false)
	if err != nil {
		t.Error(err)
	}
}

func TestUploadFolder(t *testing.T) {
	err := uploader.Upload(dir, false)
	if err == nil || !strings.HasSuffix(err.Error(), "not a Crypt4GH file") {
		t.Error(err)
	}
}

func teardown() {
	_ = os.Unsetenv("CENTRAL_EGA_USERNAME")
	_ = os.Unsetenv("CENTRAL_EGA_PASSWORD")
	_ = os.Unsetenv("LOCAL_EGA_INSTANCE_URL")
	_ = os.Unsetenv("ELIXIR_AAI_TOKEN")
}
