package scm

// MockScm is a stub implementation of the `Scm` interface to be used in our unit test suites.
// It stores the expected WorkingDir
type MockScm struct {
	Scm
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
