package dockermocks

import (
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
