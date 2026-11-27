package gitgeneric

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoteURLsSkipsRemoteWithoutURL(t *testing.T) {
	directory := t.TempDir()
	_, err := git.PlainInit(directory, false)
	require.NoError(t, err)

	config, err := os.OpenFile(filepath.Join(directory, ".git", "config"), os.O_APPEND|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	_, err = config.WriteString("[remote \"origin\"]\n\turl = https://github.com/updatecli/updatecli.git\n[remote \"nourl\"]\n\tfetch = +refs/heads/*:refs/remotes/nourl/*\n")
	require.NoError(t, err)
	require.NoError(t, config.Close())

	remotes, err := GoGit{}.RemoteURLs(directory)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"origin": "https://github.com/updatecli/updatecli.git"}, remotes)
}
