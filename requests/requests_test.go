package requests

import (
	"encoding/json"
	"github.com/logrusorgru/aurora"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var server *httptest.Server
var client Client
var url string

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := make(map[string]string)
		auth, password, _ := r.BasicAuth()
		response["auth"] = auth
		response["pass"] = password
		response["header"] = r.Header.Get("header")
		response["params"] = r.URL.RawQuery
		bytes, err := json.Marshal(response)
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
		_, err = w.Write(bytes)
		if err != nil {
			log.Fatal(aurora.Red(err))
		}
	}))
	client = NewClient(server.Client())
	url = server.URL
}

func TestDoRequest(t *testing.T) {
	username := "username"
	password := "password"
	get, err := client.DoRequest(http.MethodGet,
		url,
		strings.NewReader("Body"),
		map[string]string{"header": "test header value"},
		map[string]string{"param": "test param value"},
		username,
		password)
	if err != nil {
		t.Error(err)
	}
	body := get.Body
	defer body.Close()
	bytes, err := ioutil.ReadAll(body)
	if err != nil {
		t.Error(err)
	}
	response := string(bytes)
	if response != `{"auth":"username","header":"test header value","params":"param=test+param+value","pass":"password"}` {
		t.Error()
	}
}

func teardown() {
	server.Close()
}
