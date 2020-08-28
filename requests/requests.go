// Package requests contains structure an helper methods for performing HTTP requests towards LocalEGA instance.
package requests

import (
	"io"
	"net/http"
)

// Client is an interface providing DoRequest method for performing HTTP requests towards LocalEGA instance.
type Client interface {
	DoRequest(method string, url string, body io.Reader, headers map[string]string, params map[string]string, username string, password string) (*http.Response, error)
}

type defaultClient struct {
	client http.Client
}

// NewClient constructs Client instance, possibly accepting custom http.Client implementation.
func NewClient(client *http.Client) Client {
	defaultClient := defaultClient{}
	if client != nil {
		defaultClient.client = *client
	} else {
		defaultClient.client = *http.DefaultClient
	}
	return defaultClient
}

// DoRequest method can perform different HTTP requests with different parameters towards LocalEGA instance.
func (c defaultClient) DoRequest(method string, url string, body io.Reader, headers map[string]string, params map[string]string, username string, password string) (*http.Response, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	for name, header := range headers {
		request.Header.Add(name, header)
	}
	if params != nil {
		query := request.URL.Query()
		for name, param := range params {
			query.Add(name, param)
		}
		request.URL.RawQuery = query.Encode()
	}

	request.SetBasicAuth(username, password)

	return c.client.Do(request)
}
