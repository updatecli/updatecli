package scm

// MockScm is a stub implementation of the `ScmHandler` interface to be used in our unit test suites.
type MockScm struct {
	ScmHandler
	WorkingDir   string
	ChangedFiles []string
	Err          error
}

func (m *MockScm) GetDirectory() (directory string) {
	return m.WorkingDir
}

func (m *MockScm) GetChangedFiles(workingDir string) ([]string, error) {
	m.WorkingDir = workingDir
	return m.ChangedFiles, m.Err
}
