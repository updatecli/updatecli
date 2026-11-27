package scm

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/log"
)

func TestRepositoryKey(t *testing.T) {
	sameRepository := []string{
		"https://github.com/updatecli/updatecli.git",
		"https://github.com/updatecli/updatecli",
		"https://github.com/updatecli/updatecli/",
		"https://user:token@github.com/updatecli/updatecli.git",
		"http://GitHub.com/Updatecli/Updatecli",
		"git@github.com:updatecli/updatecli.git",
		"ssh://git@github.com:22/updatecli/updatecli.git",
		"ssh://git@github.com/updatecli/updatecli",
	}

	for _, url := range sameRepository {
		assert.Equal(t, "github.com/updatecli/updatecli", repositoryKey(url), url)
	}

	assert.NotEqual(t, repositoryKey("https://github.com/updatecli/updatecli.git"), repositoryKey("https://github.com/olblak/updatecli.git"), "another owner")
	assert.NotEqual(t, repositoryKey("https://github.com/updatecli/updatecli.git"), repositoryKey("https://gitlab.com/updatecli/updatecli.git"), "another host")
	assert.Empty(t, repositoryKey(""))
	assert.Empty(t, repositoryKey("  "))

	for _, url := range []string{
		"https://dev.azure.com/org/project/_git/repo",
		"https://org@dev.azure.com/org/project/_git/repo",
		"git@ssh.dev.azure.com:v3/org/project/repo",
	} {
		assert.Equal(t, "dev.azure.com/org/project/repo", repositoryKey(url), url)
	}
}

// urlScm is an scm handler only describing its repository URL and branch.
type urlScm struct {
	ScmHandler
	url    string
	branch string
}

func (u urlScm) GetURL() string { return u.url }

func (u urlScm) GetBranches() (string, string, string) { return u.branch, u.branch, u.branch }

// captureLog runs UsesLocalCheckout from a clean state and returns what it logged.
func captureLog(t *testing.T) *bytes.Buffer {
	t.Helper()

	resetLocalCheckout()
	t.Cleanup(resetLocalCheckout)

	var out bytes.Buffer
	previous := log.SetOutput(&out)
	t.Cleanup(func() { log.SetOutput(previous) })

	return &out
}

func newCheckout(t *testing.T, origin string) string {
	t.Helper()

	directory := t.TempDir()
	repository, err := git.PlainInit(directory, false)
	require.NoError(t, err)

	if origin != "" {
		_, err = repository.CreateRemote(&gitconfig.RemoteConfig{Name: "origin", URLs: []string{origin}})
		require.NoError(t, err)
	}

	return directory
}

func TestUsesLocalCheckoutAgainstOrigin(t *testing.T) {
	out := captureLog(t)
	directory := newCheckout(t, "git@github.com:updatecli/updatecli.git")

	assert.True(t, UsesLocalCheckout(urlScm{url: "https://github.com/updatecli/updatecli.git"}, directory),
		"the https form of the origin remote uses the local checkout")
	assert.Contains(t, out.String(), "uses the local checkout")

	other := urlScm{url: "https://user:secret@github.com/updatecli/website.git"}
	assert.False(t, UsesLocalCheckout(other, directory), "another repository is cloned")
	assert.False(t, UsesLocalCheckout(other, directory))

	assert.Equal(t, 1, strings.Count(out.String(), "it is cloned, read only"), "each url is reported once")
	assert.Contains(t, out.String(), "updatecli/website")
	assert.NotContains(t, out.String(), "secret", "credentials are redacted")
}

func TestUsesLocalCheckoutInAFork(t *testing.T) {
	out := captureLog(t)
	directory := newCheckout(t, "git@github.com:olblak/updatecli.git")

	repository, err := git.PlainOpen(directory)
	require.NoError(t, err)
	_, err = repository.CreateRemote(&gitconfig.RemoteConfig{Name: "upstream", URLs: []string{"git@github.com:updatecli/updatecli.git"}})
	require.NoError(t, err)

	assert.True(t, UsesLocalCheckout(urlScm{url: "https://github.com/updatecli/updatecli.git"}, directory),
		"the upstream repository of a fork uses the local checkout")
	assert.Contains(t, out.String(), `as its remote "upstream"`)

	assert.True(t, UsesLocalCheckout(urlScm{url: "https://github.com/olblak/updatecli.git"}, directory),
		"the fork itself uses the local checkout too")
	assert.Contains(t, out.String(), `as its remote "origin"`)

	assert.False(t, UsesLocalCheckout(urlScm{url: "https://github.com/updatecli/website.git"}, directory))
	assert.Contains(t, out.String(), "matches no remote of the local checkout")
}

func TestRemoteKeysPreferOrigin(t *testing.T) {
	directory := newCheckout(t, "https://github.com/updatecli/updatecli.git")

	repository, err := git.PlainOpen(directory)
	require.NoError(t, err)
	_, err = repository.CreateRemote(&gitconfig.RemoteConfig{Name: "aaa", URLs: []string{"git@github.com:updatecli/updatecli.git"}})
	require.NoError(t, err)
	_, err = repository.CreateRemote(&gitconfig.RemoteConfig{Name: "upstream", URLs: []string{"git@github.com:updatecli/website.git"}})
	require.NoError(t, err)

	assert.Equal(t, map[string]string{
		"github.com/updatecli/updatecli": "origin",
		"github.com/updatecli/website":   "upstream",
	}, remoteKeys(directory), "origin wins over another remote of the same repository")
}

func TestUsesLocalCheckoutWithoutRemote(t *testing.T) {
	out := captureLog(t)
	directory := newCheckout(t, "")

	assert.False(t, UsesLocalCheckout(urlScm{url: "https://github.com/updatecli/updatecli.git"}, directory),
		"without remote, nothing tells repositories apart: every scm is cloned")
	assert.False(t, UsesLocalCheckout(urlScm{url: "https://github.com/updatecli/website.git"}, directory))
	assert.Equal(t, 1, strings.Count(out.String(), "has no git remote"), "the missing remote is reported once")
}

func TestUsesLocalCheckoutInALinkedWorktree(t *testing.T) {
	gitBinary, err := exec.LookPath("git")
	if err != nil {
		t.Skip("git is needed to create a linked worktree")
	}

	out := captureLog(t)
	directory := newCheckout(t, "git@github.com:updatecli/updatecli.git")

	repository, err := git.PlainOpen(directory)
	require.NoError(t, err)
	worktree, err := repository.Worktree()
	require.NoError(t, err)
	_, err = worktree.Commit("init", &git.CommitOptions{AllowEmptyCommits: true, Author: &object.Signature{Name: "test", Email: "test@example.com"}})
	require.NoError(t, err)

	linked := filepath.Join(t.TempDir(), "linked")
	command := exec.Command(gitBinary, "-C", directory, "worktree", "add", "-q", "-b", "feat/new-manifest", linked)
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))

	assert.True(t, UsesLocalCheckout(urlScm{url: "https://github.com/updatecli/updatecli.git"}, linked),
		"the remotes of a linked worktree are read from the common git directory")
	assert.False(t, UsesLocalCheckout(urlScm{url: "https://github.com/updatecli/website.git"}, linked))
	assert.NotContains(t, out.String(), "has no git remote")
}

func TestUsesLocalCheckoutWithoutURL(t *testing.T) {
	out := captureLog(t)
	directory := newCheckout(t, "https://github.com/updatecli/updatecli.git")

	assert.True(t, UsesLocalCheckout(urlScm{}, directory), "an scm without url has nothing to clone")
	assert.Empty(t, out.String())
}

// captureDebugLog is captureLog with debug messages logged too.
func captureDebugLog(t *testing.T) *bytes.Buffer {
	t.Helper()

	out := captureLog(t)

	previous := log.GetLevel()
	log.SetLevel(log.DebugLevel)
	t.Cleanup(func() { log.SetLevel(previous) })

	return out
}

// checkoutBranch points the HEAD of the checkout in directory at branch, or detaches it.
func checkoutBranch(t *testing.T, directory, branch string) {
	t.Helper()

	repository, err := git.PlainOpen(directory)
	require.NoError(t, err)

	head := plumbing.NewSymbolicReference(plumbing.HEAD, plumbing.NewBranchReferenceName(branch))
	if branch == "" {
		head = plumbing.NewHashReference(plumbing.HEAD, plumbing.NewHash("0123456789abcdef0123456789abcdef01234567"))
	}

	require.NoError(t, repository.Storer.SetReference(head))
}

const branchMismatch = "but the local checkout"

func TestUsesLocalCheckoutSameBranch(t *testing.T) {
	out := captureDebugLog(t)
	directory := newCheckout(t, "https://github.com/updatecli/updatecli.git")
	checkoutBranch(t, directory, "main")

	UsesLocalCheckout(urlScm{url: "https://github.com/updatecli/updatecli.git", branch: "main"}, directory)
	assert.NotContains(t, out.String(), "targets branch")
}

func TestUsesLocalCheckoutOtherBranch(t *testing.T) {
	out := captureDebugLog(t)
	directory := newCheckout(t, "https://github.com/updatecli/updatecli.git")
	checkoutBranch(t, directory, "feat/new-manifest")

	main := urlScm{url: "https://github.com/updatecli/updatecli.git", branch: "main"}
	assert.True(t, UsesLocalCheckout(main, directory), "another branch of the same repository still uses the local checkout")
	UsesLocalCheckout(main, directory)

	assert.Equal(t, 1, strings.Count(out.String(), `targets branch "main"`), "each scm and branch is logged once")
	assert.Contains(t, out.String(), `is on branch "feat/new-manifest"`)
	assert.Contains(t, out.String(), "DEBUG")
	assert.NotContains(t, out.String(), "WARN")

	UsesLocalCheckout(urlScm{url: "https://github.com/updatecli/updatecli.git", branch: "release/1.x"}, directory)
	assert.Equal(t, 2, strings.Count(out.String(), branchMismatch), "another branch of the same scm is logged too")
}

func TestUsesLocalCheckoutDetachedHead(t *testing.T) {
	out := captureDebugLog(t)
	directory := newCheckout(t, "https://github.com/updatecli/updatecli.git")
	checkoutBranch(t, directory, "")

	UsesLocalCheckout(urlScm{url: "https://github.com/updatecli/updatecli.git", branch: "main"}, directory)
	assert.Contains(t, out.String(), "is not on a branch")
}

func TestUsesLocalCheckoutBranchOnlyForTheLocalCheckout(t *testing.T) {
	out := captureDebugLog(t)
	directory := newCheckout(t, "https://github.com/updatecli/updatecli.git")
	checkoutBranch(t, directory, "feat/new-manifest")

	assert.False(t, UsesLocalCheckout(urlScm{url: "https://github.com/updatecli/website.git", branch: "main"}, directory))
	assert.NotContains(t, out.String(), "targets branch", "a cloned scm uses its own branch")
}
