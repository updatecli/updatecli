package gitgeneric

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckoutPullsBranchOnce(t *testing.T) {
	remoteDirectory := t.TempDir()
	remote, err := git.PlainInit(remoteDirectory, false)
	require.NoError(t, err)

	remoteWorktree, err := remote.Worktree()
	require.NoError(t, err)

	commitToRemote := func(content string) plumbing.Hash {
		require.NoError(t, os.WriteFile(filepath.Join(remoteDirectory, "file.txt"), []byte(content), 0o600))
		_, err := remoteWorktree.Add("file.txt")
		require.NoError(t, err)
		hash, err := remoteWorktree.Commit(content, &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Updatecli Test",
				Email: "test@updatecli.io",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)
		return hash
	}

	initialHash := commitToRemote("initial")

	localDirectory := t.TempDir()
	local, err := git.PlainClone(localDirectory, false, &git.CloneOptions{URL: remoteDirectory})
	require.NoError(t, err)

	localHead := func() plumbing.Hash {
		head, err := local.Head()
		require.NoError(t, err)
		return head.Hash()
	}
	require.Equal(t, initialHash, localHead())

	g := GoGit{}

	firstRemoteHash := commitToRemote("first")
	require.NoError(t, g.Checkout("", "", "master", "master", localDirectory, false, nil))
	assert.Equal(t, firstRemoteHash, localHead(), "the first checkout of a branch should pull it")

	commitToRemote("second")
	require.NoError(t, g.Checkout("", "", "master", "master", localDirectory, false, nil))
	assert.Equal(t, firstRemoteHash, localHead(), "later checkouts of a branch should not pull it again")
}
