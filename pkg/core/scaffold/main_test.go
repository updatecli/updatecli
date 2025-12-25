package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {

	testRootDir := filepath.Join(os.TempDir(), "updatecli", "test", "core", "scaffold")

	s := Scaffold{}

	err := s.Run(testRootDir)
	require.NoError(t, err)

	assert.DirExists(t, filepath.Join(testRootDir))
	assert.FileExists(t, filepath.Join(testRootDir, "Policy.yaml"))

	assert.DirExists(t, filepath.Join(testRootDir, "updatecli.d"))
	assert.FileExists(t, filepath.Join(testRootDir, "updatecli.d", "_scm.bitbucket.yaml"))
	assert.FileExists(t, filepath.Join(testRootDir, "updatecli.d", "_scm.github.yaml"))
	assert.FileExists(t, filepath.Join(testRootDir, "updatecli.d", "_scm.gitea.yaml"))
	assert.FileExists(t, filepath.Join(testRootDir, "updatecli.d", "_scm.gitlab.yaml"))
	assert.FileExists(t, filepath.Join(testRootDir, "updatecli.d", "_scm.stash.yaml"))
	assert.FileExists(t, filepath.Join(testRootDir, "updatecli.d", "default.example.yaml"))

	assert.FileExists(t, filepath.Join(testRootDir, "README.md"))

	assert.FileExists(t, filepath.Join(testRootDir, "values.yaml"))

	assert.FileExists(t, filepath.Join(testRootDir, "CHANGELOG.md"))

	// Cleanup after test
	os.RemoveAll(testRootDir)
}
