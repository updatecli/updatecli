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

func TestCloneMirrorsBranchesAndTagsOnly(t *testing.T) {
	remoteDirectory := t.TempDir()
	remote, err := git.PlainInit(remoteDirectory, false)
	require.NoError(t, err)

	worktree, err := remote.Worktree()
	require.NoError(t, err)

	commit := func(content string) plumbing.Hash {
		require.NoError(t, os.WriteFile(filepath.Join(remoteDirectory, "file.txt"), []byte(content), 0o600))
		_, err := worktree.Add("file.txt")
		require.NoError(t, err)
		hash, err := worktree.Commit(content, &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Updatecli Test",
				Email: "test@updatecli.io",
				When:  time.Now(),
			},
		})
		require.NoError(t, err)
		return hash
	}

	mainHash := commit("main")
	_, err = remote.CreateTag("v1.0.0", mainHash, nil)
	require.NoError(t, err)
	require.NoError(t, remote.Storer.SetReference(
		plumbing.NewHashReference(plumbing.NewBranchReferenceName("updatecli_main_pipeline"), mainHash)))

	// GitHub keeps such a ref for every pull request
	pullRequestHash := commit("pull request")
	require.NoError(t, remote.Storer.SetReference(
		plumbing.NewHashReference(plumbing.ReferenceName("refs/pull/1/head"), pullRequestHash)))
	require.NoError(t, remote.Storer.SetReference(
		plumbing.NewHashReference(plumbing.NewBranchReferenceName("master"), mainHash)))

	localDirectory := t.TempDir()
	require.NoError(t, GoGit{}.Clone("", "", remoteDirectory, localDirectory, nil, nil, "master", false))

	local, err := git.PlainOpen(localDirectory)
	require.NoError(t, err)

	for _, name := range []plumbing.ReferenceName{
		plumbing.NewBranchReferenceName("updatecli_main_pipeline"),
		plumbing.NewTagReferenceName("v1.0.0"),
	} {
		_, err := local.Reference(name, false)
		assert.NoError(t, err, "%s should be mirrored", name)
	}

	_, err = local.Reference("refs/pull/1/head", false)
	assert.ErrorIs(t, err, plumbing.ErrReferenceNotFound, "pull request refs should not be fetched")
}
