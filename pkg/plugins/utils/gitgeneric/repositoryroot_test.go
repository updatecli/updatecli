package gitgeneric

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepositoryRoot(t *testing.T) {
	root := t.TempDir()
	_, err := git.PlainInit(root, false)
	require.NoError(t, err)

	subDir := filepath.Join(root, "updatecli", "updatecli.d")
	require.NoError(t, os.MkdirAll(subDir, 0o750))

	for _, path := range []string{root, subDir} {
		directory, err := GoGit{}.RepositoryRoot(path)
		require.NoError(t, err, path)
		assert.Equal(t, root, directory, path)
	}

	_, err = GoGit{}.RepositoryRoot(t.TempDir())
	assert.Error(t, err, "a directory outside any git repository has no root")
}
