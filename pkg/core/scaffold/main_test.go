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
	assert.FileExists(t, filepath.Join(testRootDir, "updatecli.d", "default.yaml"))

	assert.FileExists(t, filepath.Join(testRootDir, "README.md"))

	assert.DirExists(t, filepath.Join(testRootDir, "values.d"))
	assert.FileExists(t, filepath.Join(testRootDir, "values.d", "default.yaml"))

	assert.FileExists(t, filepath.Join(testRootDir, "CHANGELOG.md"))

	// Cleanup after test
	os.RemoveAll(testRootDir)
}
