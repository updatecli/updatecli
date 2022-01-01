package dockermocks

import (
	"net/http"

	"github.com/updatecli/updatecli/pkg/plugins/docker/dockerimage"
)

// MockRegistry is a stub implementation of the `Registry` interface to be used in our unit test suites.
type MockRegistry struct {
	ReturnedDigest string
	ReturnedError  error
	InputImageName string
}

func (m *MockRegistry) Digest(image dockerimage.Image) (string, error) {
	m.InputImageName = image.FullName()
	return m.ReturnedDigest, m.ReturnedError
}

// MockClient is a stub implementation of an http.Client with a custom Do method
type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

// Do is the mock client's `Do` func that you can customize
func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}
