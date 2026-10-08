package gitgeneric

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	"github.com/go-git/go-git/v5/plumbing/transport/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const workingBranch = "updatecli_main_test"

// singleBranchRemote is a bare git repository used as a remote, alongside a
// separate clone used to push commits to it, like another actor would do.
type singleBranchRemote struct {
	t    *testing.T
	url  string
	seed string
}

// gitCmd runs a git command in dir, ignoring the git configuration of the machine running the tests,
// and returns its combined output. The test fails if the command fails.
func gitCmd(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "commit.gpgsign=false"}, args...)...)
	cmd.Dir = dir
	// Ignore the git configuration of the machine running the tests
	cmd.Env = append(os.Environ(),
		"GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_SYSTEM="+os.DevNull, "GIT_CONFIG_NOSYSTEM=1",
		"GIT_AUTHOR_NAME=test", "GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test", "GIT_COMMITTER_EMAIL=test@example.com")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v: %s", args, out)
	return string(out)
}

// newSingleBranchRemote creates a bare repository, containing a single commit on the main branch,
// and the clone used to push to it.
func newSingleBranchRemote(t *testing.T) *singleBranchRemote {
	t.Helper()
	root := t.TempDir()
	bare := filepath.Join(root, "remote.git")
	seed := filepath.Join(root, "seed")
	require.NoError(t, os.MkdirAll(seed, 0o755))

	gitCmd(t, root, "init", "-q", "--bare", "-b", "main", bare)
	gitCmd(t, seed, "init", "-q", "-b", "main")
	gitCmd(t, seed, "remote", "add", "origin", bare)

	r := &singleBranchRemote{t: t, url: "file://" + bare, seed: seed}
	r.commit("main", "main-1")
	gitCmd(t, seed, "push", "-q", "origin", "main")
	return r
}

// commit creates a commit on the given branch of the seed clone, creating the
// branch from the current HEAD if needed, and pushes it.
func (r *singleBranchRemote) commit(branch, file string) {
	r.t.Helper()
	gitCmd(r.t, r.seed, "fetch", "-q", "origin")
	switch {
	case gitCmd(r.t, r.seed, "ls-remote", "--heads", "origin", branch) != "":
		// the branch may have been updated by someone else since the last commit
		gitCmd(r.t, r.seed, "checkout", "-q", "-B", branch, "origin/"+branch)
	case gitCmd(r.t, r.seed, "branch", "--list", branch) != "":
		gitCmd(r.t, r.seed, "checkout", "-q", branch)
	default:
		gitCmd(r.t, r.seed, "checkout", "-q", "-b", branch)
	}
	require.NoError(r.t, os.WriteFile(filepath.Join(r.seed, file), []byte(file), 0o644))
	gitCmd(r.t, r.seed, "add", file)
	gitCmd(r.t, r.seed, "commit", "-qm", file)
	gitCmd(r.t, r.seed, "push", "-q", "origin", branch)
}

// head returns the latest commit hash of a branch in the seed clone, which is the remote one
// once the branch is pushed by commit.
func (r *singleBranchRemote) head(branch string) string {
	r.t.Helper()
	return strings.TrimSpace(gitCmd(r.t, r.seed, "rev-parse", branch))
}

// TestSingleBranchMatchesFullClone checks that enabling singleBranch does not
// change how an already published working branch is handled by Checkout and
// Pull, compared to a regular clone. Pull is run after another actor pushed a
// commit to the working branch, as it happens when commits are created
// using the GitHub API.
func TestSingleBranchMatchesFullClone(t *testing.T) {
	tests := []struct {
		name string
		// setup prepares the remote before updatecli runs.
		setup func(r *singleBranchRemote, branch string)
		// skipShallow skips the depth 1 run. Detecting that a branch diverged requires finding a
		// common ancestor, which a depth 1 history doesn't have, even with a regular clone.
		skipShallow bool
	}{
		{
			name:  "working branch does not exist yet",
			setup: func(r *singleBranchRemote, branch string) {},
		},
		{
			name: "working branch is ahead of main",
			setup: func(r *singleBranchRemote, branch string) {
				r.commit(branch, "wb-1")
				gitCmd(t, r.seed, "checkout", "-q", "main")
			},
		},
		{
			name: "working branch diverged from main",
			setup: func(r *singleBranchRemote, branch string) {
				r.commit(branch, "wb-1")
				r.commit("main", "main-2")
			},
			skipShallow: true,
		},
	}

	type outcome struct {
		localAfterCheckout string
		forceReset         bool
		localAfterPull     string
	}

	run := func(t *testing.T, r *singleBranchRemote, branch string, singleBranch bool, depth *int) outcome {
		dir := filepath.Join(t.TempDir(), "clone")
		g := &GoGit{}

		require.NoError(t, g.Clone("", "", r.url, dir, nil, depth, "main", singleBranch))
		require.NoError(t, g.Checkout("", "", "main", branch, dir, true, depth))

		var o outcome
		var err error
		o.localAfterCheckout, err = g.GetLatestCommitHash(dir)
		require.NoError(t, err)
		o.forceReset = g.IsForceReset()

		// Simulate the commit created remotely, for instance with the GitHub API
		if g.IsForceReset() || o.localAfterCheckout == r.head("main") {
			// the working branch is based on main, push it first like updatecli does
			_, err = g.Push("", "", dir, true)
			require.NoError(t, err)
		}
		r.commit(branch, "api-commit")

		require.NoError(t, g.Pull("", "", dir, branch, singleBranch, true, depth))
		o.localAfterPull, err = g.GetLatestCommitHash(dir)
		require.NoError(t, err)
		assert.Equal(t, r.head(branch), o.localAfterPull, "local branch should match the remote after Pull")
		return o
	}

	one := 1
	// working branch names that are valid refs but not trivial refspec components
	branches := []string{workingBranch, "updatecli/main/abc", "updatecli.v1.2.x"}

	for _, branch := range branches {
		for _, depth := range []*int{nil, &one} {
			depthName := "full history"
			if depth != nil {
				depthName = "depth 1"
			}

			for _, tt := range tests {
				if depth != nil && tt.skipShallow {
					continue
				}

				t.Run(branch+"/"+depthName+"/"+tt.name, func(t *testing.T) {
					// each mode gets its own identical remote
					rFull := newSingleBranchRemote(t)
					tt.setup(rFull, branch)
					full := run(t, rFull, branch, false, depth)

					rSingle := newSingleBranchRemote(t)
					tt.setup(rSingle, branch)
					single := run(t, rSingle, branch, true, depth)

					assert.Equal(t, full.forceReset, single.forceReset, "ForceReset should not depend on singleBranch")
					assert.Equal(t,
						full.localAfterCheckout == rFull.head("main"),
						single.localAfterCheckout == rSingle.head("main"),
						"working branch should start from the same base regardless of singleBranch")
				})
			}
		}
	}
}

// TestCheckoutWithUnreachableRemote checks that a new working branch is still created
// from the base branch when the remote can't be reached to look for an existing one.
func TestCheckoutWithUnreachableRemote(t *testing.T) {
	r := newSingleBranchRemote(t)
	dir := filepath.Join(t.TempDir(), "clone")
	g := &GoGit{}

	require.NoError(t, g.Clone("", "", r.url, dir, nil, nil, "main", true))
	gitCmd(t, dir, "remote", "set-url", "origin", "file://"+filepath.Join(t.TempDir(), "missing.git"))

	require.NoError(t, g.Checkout("", "", "main", workingBranch, dir, true, nil))

	hash, err := g.GetLatestCommitHash(dir)
	require.NoError(t, err)
	assert.Equal(t, r.head("main"), hash)
}

// racyTransport wraps the local git transport to run a hook right before the Nth
// upload-pack session started after being armed. The remote references are computed when
// a session starts, so the hook lets a test update the remote between two requests of
// the same git operation.
type racyTransport struct {
	transport.Transport

	mu       sync.Mutex
	sessions int
	runOn    int
	hook     func()
}

// racyScheme is the URL scheme registered for racyTransport.
const racyScheme = "racefile"

var (
	racy         = &racyTransport{Transport: file.DefaultClient}
	registerRacy sync.Once
)

// arm runs hook right before the nth upload-pack session started from now on.
func (r *racyTransport) arm(nth int, hook func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions, r.runOn, r.hook = 0, nth, hook
}

// NewUploadPackSession implements transport.Transport.
func (r *racyTransport) NewUploadPackSession(ep *transport.Endpoint, auth transport.AuthMethod) (transport.UploadPackSession, error) {
	r.mu.Lock()
	r.sessions++
	var hook func()
	if r.hook != nil && r.sessions == r.runOn {
		hook, r.hook = r.hook, nil
	}
	r.mu.Unlock()

	if hook != nil {
		hook()
	}

	return r.Transport.NewUploadPackSession(ep, auth)
}

// TestPullWorkingBranchUpdatedDuringPull checks that Pull still succeeds when the working
// branch is updated on the remote while the pull is running. The branch must be fetched
// by the same request that resolves its latest commit, otherwise the commit is missing locally.
func TestPullWorkingBranchUpdatedDuringPull(t *testing.T) {
	registerRacy.Do(func() { client.InstallProtocol(racyScheme, racy) })

	r := newSingleBranchRemote(t)
	r.commit(workingBranch, "wb-1")
	gitCmd(t, r.seed, "checkout", "-q", "main")

	url := racyScheme + strings.TrimPrefix(r.url, "file")
	dir := filepath.Join(t.TempDir(), "clone")
	g := &GoGit{}

	require.NoError(t, g.Clone("", "", url, dir, nil, nil, "main", true))
	require.NoError(t, g.Checkout("", "", "main", workingBranch, dir, true, nil))

	// commit created remotely, for instance with the GitHub API
	r.commit(workingBranch, "api-commit")

	// someone else updates the working branch before the second request of the pull
	racy.arm(2, func() { r.commit(workingBranch, "concurrent-push") })
	t.Cleanup(func() { racy.arm(0, nil) })

	require.NoError(t, g.Pull("", "", dir, workingBranch, true, true, nil))
}

// remoteFetchRefSpecs returns the fetch refspecs configured for the origin remote.
func remoteFetchRefSpecs(t *testing.T, dir string) []string {
	t.Helper()
	return strings.Fields(gitCmd(t, dir, "config", "--get-all", "remote.origin.fetch"))
}

// TestPullRestoresRemoteConfiguration checks that the working branch is only fetched for the
// duration of a pull, so a later fetch doesn't depend on the branch still existing on the remote.
func TestPullRestoresRemoteConfiguration(t *testing.T) {
	r := newSingleBranchRemote(t)
	r.commit(workingBranch, "wb-1")
	gitCmd(t, r.seed, "checkout", "-q", "main")

	dir := filepath.Join(t.TempDir(), "clone")
	g := &GoGit{}
	require.NoError(t, g.Clone("", "", r.url, dir, nil, nil, "main", true))
	require.NoError(t, g.Checkout("", "", "main", workingBranch, dir, true, nil))

	before := remoteFetchRefSpecs(t, dir)
	require.Equal(t, []string{"+refs/heads/main:refs/remotes/origin/main"}, before)

	require.NoError(t, g.Pull("", "", dir, workingBranch, true, true, nil))
	assert.Equal(t, before, remoteFetchRefSpecs(t, dir), "remote configuration should be restored after a successful pull")

	// the branch is deleted from the remote, the pull fails and the configuration is restored again
	gitCmd(t, r.seed, "push", "-q", "origin", "--delete", workingBranch)
	assert.Error(t, g.Pull("", "", dir, workingBranch, true, true, nil))
	assert.Equal(t, before, remoteFetchRefSpecs(t, dir), "remote configuration should be restored after a failed pull")
}

// TestSingleBranchWorkingBranchIsBaseBranch checks the case where no separate working branch
// is used, so the branch to check out and pull is the branch that was cloned.
func TestSingleBranchWorkingBranchIsBaseBranch(t *testing.T) {
	r := newSingleBranchRemote(t)
	dir := filepath.Join(t.TempDir(), "clone")
	g := &GoGit{}

	require.NoError(t, g.Clone("", "", r.url, dir, nil, nil, "main", true))
	require.NoError(t, g.Checkout("", "", "main", "main", dir, true, nil))
	assert.False(t, g.IsForceReset())

	r.commit("main", "main-2")

	require.NoError(t, g.Pull("", "", dir, "main", true, true, nil))

	hash, err := g.GetLatestCommitHash(dir)
	require.NoError(t, err)
	assert.Equal(t, r.head("main"), hash)
}
