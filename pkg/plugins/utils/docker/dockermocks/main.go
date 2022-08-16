package dockermocks

import (
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker/dockerimage"
)

// MockRegistry is a stub implementation of the `Registry` interface to be used in our unit test suites.
type MockRegistry struct {
	ReturnedTags   []string
	ReturnedDigest string
	ReturnedError  error
	InputImageName string
}

func (m *MockRegistry) Digest(image dockerimage.Image) (string, error) {
	m.InputImageName = image.FullName()
	return m.ReturnedDigest, m.ReturnedError
}

func (m *MockRegistry) Tags(image dockerimage.Image) ([]string, error) {
	return m.ReturnedTags, m.ReturnedError
}
