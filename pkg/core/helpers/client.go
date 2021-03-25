package helpers

import (
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HttpClient interface to wrap Do to make testing easier.
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type DefaultHttpClient struct {

}

func (d *DefaultHttpClient) Do(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

type FakeHttpClient struct {
	Requests map[string]FakeResponse
}

type FakeResponse struct {
	StatusCode int
	Body string
	Headers map[string][]string
}

func (d *FakeHttpClient) Do(req *http.Request) (*http.Response, error) {
	if response, ok := d.Requests[req.URL.String()]; ok {
		res := http.Response{
			StatusCode: response.StatusCode,
			Body: io.NopCloser(strings.NewReader(response.Body)),
			Header: response.Headers,
		}
		return &res, nil
	} else {
		res, err := http.DefaultClient.Do(req)
		fmt.Printf("response = %+v\n", res)
		return res, err
	}
}
