package gitgeneric

// MockGit is a stub implementation of the `GitHandler` interface to be used in our unit test suites.
type MockGit struct {
	GitHandler // Ensure any unspecified method from this interface are still declared
	Remotes    map[string]string
	Err        error
}

func (m MockGit) RemoteURLs(workingDir string) (map[string]string, error) {
	return m.Remotes, m.Err
}
