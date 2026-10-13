package gittag

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
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// newTaggedRepository returns a repository with a lightweight tag "v1.0.0" and an annotated
// tag "v2.0.0", and the commits they point to.
func newTaggedRepository(t *testing.T) (dir string, lightweightCommit, annotatedCommit plumbing.Hash) {
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

	lightweightCommit = commit("first")
	_, err = repository.CreateTag("v1.0.0", lightweightCommit, nil)
	require.NoError(t, err)

	annotatedCommit = commit("second")
	_, err = repository.CreateTag("v2.0.0", annotatedCommit, &git.CreateTagOptions{Tagger: signature, Message: "v2.0.0"})
	require.NoError(t, err)

	return dir, lightweightCommit, annotatedCommit
}

func TestListRemoteURLTagsMatchesClonedTags(t *testing.T) {
	t.Setenv("TMPDIR", t.TempDir())
	remote, lightweightCommit, annotatedCommit := newTaggedRepository(t)

	gt := GitTag{spec: Spec{URL: remote}, nativeGitHandler: &gitgeneric.GoGit{}}

	remoteTags, remoteHashes, err := gt.listRemoteURLTags()
	require.NoError(t, err)
	clonedTags, clonedHashes, err := gt.listRemoteDirectoryTags("", age.Spec{}, pathresolver.Resolver{})
	require.NoError(t, err)

	assert.ElementsMatch(t, clonedTags, remoteTags)
	// A pinned reference needs the commit, not the tag object of an annotated tag.
	assert.Equal(t, map[string]string{
		"v1.0.0": lightweightCommit.String(),
		"v2.0.0": annotatedCommit.String(),
	}, remoteHashes)
	assert.Equal(t, clonedHashes, remoteHashes)
}
