package mavenmetadata

// MockMetadataHandler implements the MetadataHandler interface to provide a mock
// to be used for unit tests
type MockMetadataHandler struct {
	LatestVersion string
	Versions      []string
	Err           error
}

func (m *MockMetadataHandler) GetLatestVersion() (string, error) {
	return m.LatestVersion, m.Err
}

func (m *MockMetadataHandler) GetVersions() ([]string, error) {
	return m.Versions, m.Err
}

func (m *MockMetadataHandler) GetMetadataURL() string {
	return "It's a mock"
}
