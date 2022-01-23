package httpclient

import "net/http"

// MockClient is a stub implementation of an http.Client with a custom Do method
type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// Do is the mock client's `Do` func that you can customize
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}
