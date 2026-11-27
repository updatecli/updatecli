package scm

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

// stubScm records the calls a local checkout must never forward.
type stubScm struct {
	ScmHandler
	calls []string
}

func (s *stubScm) GetBranches() (string, string, string) { return "main", "updatecli_main", "main" }
func (s *stubScm) GetURL() string                        { return "https://example.com/org/repo.git" }
func (s *stubScm) Summary() string                       { return "example.com/org/repo" }

func (s *stubScm) Clone() (string, error) { s.calls = append(s.calls, "clone"); return "", nil }
func (s *stubScm) Checkout() error        { s.calls = append(s.calls, "checkout"); return nil }
func (s *stubScm) Clean() error           { s.calls = append(s.calls, "clean"); return nil }
func (s *stubScm) Push() (bool, error)    { s.calls = append(s.calls, "push"); return false, nil }

func TestLocalNeverTouchesTheCheckout(t *testing.T) {
	dir := t.TempDir()
	inner := &stubScm{}
	local := NewLocal(inner, dir)

	directory, err := local.Clone()
	require.NoError(t, err)
	assert.Equal(t, dir, directory)
	require.NoError(t, local.Checkout())
	require.NoError(t, local.Clean())
	require.NoError(t, local.Add([]string{"file"}))
	require.NoError(t, local.Commit(context.Background(), "message"))
	require.NoError(t, local.PushTag("v1"))
	require.NoError(t, local.PushBranch("branch"))

	reset, err := local.Push()
	require.NoError(t, err)
	assert.False(t, reset)

	cleaned, err := local.CleanWorkingBranch()
	require.NoError(t, err)
	assert.False(t, cleaned)

	upToDate, err := local.IsRemoteBranchUpToDate()
	require.NoError(t, err)
	assert.True(t, upToDate)

	exists, err := local.IsRemoteWorkingBranchExist()
	require.NoError(t, err)
	assert.True(t, exists)

	assert.Empty(t, inner.calls, "no git operation may reach the wrapped scm")
	assert.DirExists(t, dir, "the local checkout must never be removed")
}

func TestLocalDelegatesDescription(t *testing.T) {
	local := NewLocal(&stubScm{}, "/work/repo")

	source, working, target := local.GetBranches()
	assert.Equal(t, []string{"main", "updatecli_main", "main"}, []string{source, working, target})
	assert.Equal(t, "https://example.com/org/repo.git", local.GetURL())
	assert.Equal(t, "example.com/org/repo (local checkout /work/repo)", local.Summary())
	assert.Equal(t, "/work/repo", local.GetDirectory())
}

func TestLocalGetChangedFiles(t *testing.T) {
	local := NewLocal(&stubScm{}, "/work/repo")
	local.nativeGitHandler = gitgeneric.MockGit{ChangedFiles: []string{"Dockerfile"}}

	files, err := local.GetChangedFiles("/work/repo")
	require.NoError(t, err)
	assert.Equal(t, []string{"Dockerfile"}, files)
}

func TestLocalGetChangedFilesLeavesOutExistingWork(t *testing.T) {
	directory := t.TempDir()
	repository, err := git.PlainInit(directory, false)
	require.NoError(t, err)

	write := func(name, content string) {
		t.Helper()
		require.NoError(t, os.WriteFile(filepath.Join(directory, name), []byte(content), 0o600))
	}

	for _, name := range []string{"wip.txt", "edited.txt", "target.txt"} {
		write(name, "committed\n")
	}
	worktree, err := repository.Worktree()
	require.NoError(t, err)
	_, err = worktree.Add(".")
	require.NoError(t, err)
	_, err = worktree.Commit("init", &git.CommitOptions{Author: &object.Signature{Name: "test", Email: "test@example.com"}})
	require.NoError(t, err)

	// The author's own uncommitted work, before the resource runs.
	write("wip.txt", "work in progress\n")
	write("edited.txt", "work in progress\n")
	write("untracked.txt", "new\n")

	local := NewLocal(&stubScm{}, directory)
	require.NoError(t, local.Checkout())

	// What the resource changes.
	write("target.txt", "changed by the target\n")
	write("edited.txt", "changed by the target\n")

	files, err := local.GetChangedFiles(directory)
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"target.txt", "edited.txt"}, files,
		"files changed before the resource ran are left out, unless the resource changed them again")
}
