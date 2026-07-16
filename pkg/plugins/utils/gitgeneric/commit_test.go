package gitgeneric

import (
	"os"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCommitHash(t *testing.T) {
	directory := t.TempDir()
	repository, err := git.PlainInit(directory, false)
	require.NoError(t, err)

	worktree, err := repository.Worktree()
	require.NoError(t, err)

	commit := func(content, message string) plumbing.Hash {
		require.NoError(t, os.WriteFile(directory+"/file.txt", []byte(content), 0o600))
		_, err = worktree.Add("file.txt")
		require.NoError(t, err)
		hash, commitErr := worktree.Commit(message, &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Updatecli Test",
				Email: "test@updatecli.io",
				When:  time.Now(),
			},
		})
		require.NoError(t, commitErr)
		return hash
	}

	mainHash := commit("main", "main commit")
	featureReference := plumbing.NewHashReference(plumbing.NewBranchReferenceName("feature"), mainHash)
	require.NoError(t, repository.Storer.SetReference(featureReference))
	require.NoError(t, worktree.Checkout(&git.CheckoutOptions{Branch: featureReference.Name()}))
	featureHash := commit("feature", "feature commit")

	remoteReference := plumbing.NewHashReference(
		plumbing.NewRemoteReferenceName(DefaultRemoteReferenceName, "remote-only"),
		featureHash,
	)
	require.NoError(t, repository.Storer.SetReference(remoteReference))

	handler := GoGit{}
	tests := []struct {
		name       string
		branch     string
		want       string
		wantErr    string
		workingDir string
	}{
		{name: "current HEAD", want: featureHash.String(), workingDir: directory},
		{name: "local branch", branch: "master", want: mainHash.String(), workingDir: directory},
		{name: "remote branch", branch: "remote-only", want: featureHash.String(), workingDir: directory},
		{name: "missing branch", branch: "missing", wantErr: `branch "missing" not found`, workingDir: directory},
		{name: "invalid repository", branch: "master", wantErr: "opening", workingDir: directory + "/missing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handler.GetCommitHash(tt.workingDir, tt.branch)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
