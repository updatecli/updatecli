package httpclient

import "net/http"

// MockClient is a stub implementation of an http.Client with a custom `Do` method
type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// Do is the mock client's `Do` function that you can customize
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func (m *MockClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	return m.Do(req)
}
