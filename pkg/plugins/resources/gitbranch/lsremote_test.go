package gitbranch

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// newBranchedRepository returns a repository with the branches "master" and "release-1",
// and the commit "release-1" points to.
func newBranchedRepository(t *testing.T) (dir string, releaseCommit plumbing.Hash) {
	t.Helper()
	dir = t.TempDir()
	repository, err := git.PlainInit(dir, false)
	require.NoError(t, err)
	worktree, err := repository.Worktree()
	require.NoError(t, err)

	signature := &object.Signature{Name: "Updatecli Test", Email: "test@updatecli.io", When: time.Now()}
	commit := func(content string) plumbing.Hash {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "file.txt"), []byte(content), 0o600))
		_, err := worktree.Add("file.txt")
		require.NoError(t, err)
		hash, err := worktree.Commit(content, &git.CommitOptions{Author: signature})
		require.NoError(t, err)
		return hash
	}

	releaseCommit = commit("first")
	require.NoError(t, repository.Storer.SetReference(plumbing.NewHashReference(plumbing.NewBranchReferenceName("release-1"), releaseCommit)))
	commit("second")

	return dir, releaseCommit
}

func TestLsRemoteCondition(t *testing.T) {
	remote, _ := newBranchedRepository(t)

	for branch, want := range map[string]bool{"release-1": true, "master": true, "release-2": false} {
		gb, err := New(Spec{URL: remote, Branch: branch, LsRemote: new(true)})
		require.NoError(t, err)

		pass, _, err := gb.Condition(context.Background(), "", nil, pathresolver.Resolver{})
		require.NoError(t, err)
		assert.Equal(t, want, pass, branch)
	}
}

func TestLsRemoteSourceReturnsBranchCommit(t *testing.T) {
	remote, releaseCommit := newBranchedRepository(t)

	gb, err := New(Spec{
		URL:           remote,
		Key:           "hash",
		LsRemote:      new(true),
		VersionFilter: version.Filter{Kind: "regex", Pattern: "^release-1$"},
	})
	require.NoError(t, err)

	got := result.Source{}
	require.NoError(t, gb.Source(context.Background(), pathresolver.Resolver{}, &got))
	assert.Equal(t, releaseCommit.String(), got.Information)
}

func TestLsRemoteValidation(t *testing.T) {
	for name, spec := range map[string]Spec{
		"without url": {LsRemote: new(true)},
		"with path":   {URL: "https://example.com/repo.git", Path: "repo", LsRemote: new(true)},
		"with age":    {URL: "https://example.com/repo.git", Age: age.Spec{Minimum: "7d"}, LsRemote: new(true)},
	} {
		_, err := New(spec)
		assert.Error(t, err, name)
	}
}
