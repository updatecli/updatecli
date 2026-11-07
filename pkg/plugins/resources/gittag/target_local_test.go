package gittag

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
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// masterScm is the scm a local checkout wraps: only its branches are read.
type masterScm struct{ scm.ScmHandler }

func (masterScm) GetBranches() (string, string, string) { return "master", "master", "master" }

// TestTargetLeavesLocalCheckoutUntouched runs the target, outside dry run, against a local
// checkout with uncommitted and untracked work: a forced checkout used to discard both.
func TestTargetLeavesLocalCheckoutUntouched(t *testing.T) {
	directory := t.TempDir()
	repository, err := git.PlainInit(directory, false)
	require.NoError(t, err)

	notes := filepath.Join(directory, "notes.txt")
	require.NoError(t, os.WriteFile(notes, []byte("committed\n"), 0o600))
	worktree, err := repository.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add("notes.txt")
	require.NoError(t, err)
	_, err = worktree.Commit("init", &git.CommitOptions{Author: &object.Signature{Name: "t", Email: "t@t", When: time.Now()}})
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(notes, []byte("committed\nuncommitted\n"), 0o600))
	untracked := filepath.Join(directory, "updatecli.d", "manifest.yaml")
	require.NoError(t, os.MkdirAll(filepath.Dir(untracked), 0o750))
	require.NoError(t, os.WriteFile(untracked, []byte("name: test\n"), 0o600))

	head, err := repository.Head()
	require.NoError(t, err)

	spec := map[string]interface{}{}
	res, err := New(spec)
	require.NoError(t, err)

	local := scm.NewLocal(masterScm{}, directory)
	gotResult := result.Target{}
	targetErr := res.Target(context.Background(), "v1.0.0", local, pathresolver.New(local, ""), false, &gotResult)

	content, err := os.ReadFile(notes)
	require.NoError(t, err)
	assert.Equal(t, "committed\nuncommitted\n", string(content), "uncommitted work is kept")
	assert.FileExists(t, untracked, "untracked files are kept")

	require.NoError(t, targetErr)
	assert.True(t, gotResult.Changed, "the gittag to create is reported")
	assert.Equal(t, result.ATTENTION, gotResult.Result)

	newHead, err := repository.Head()
	require.NoError(t, err)
	assert.Equal(t, head.Name(), newHead.Name(), "the current branch is kept")

	tags, err := repository.Tags()
	require.NoError(t, err)
	count := 0
	require.NoError(t, tags.ForEach(func(*plumbing.Reference) error { count++; return nil }))
	assert.Equal(t, 0, count, "no git tag is created in the local checkout")
}
