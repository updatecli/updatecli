package gitgeneric

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

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
		setup func(r *singleBranchRemote)
		// skipShallow skips the depth 1 run. Detecting that a branch diverged requires finding a
		// common ancestor, which a depth 1 history doesn't have, even with a regular clone.
		skipShallow bool
	}{
		{
			name:  "working branch does not exist yet",
			setup: func(r *singleBranchRemote) {},
		},
		{
			name: "working branch is ahead of main",
			setup: func(r *singleBranchRemote) {
				r.commit(workingBranch, "wb-1")
				gitCmd(t, r.seed, "checkout", "-q", "main")
			},
		},
		{
			name: "working branch diverged from main",
			setup: func(r *singleBranchRemote) {
				r.commit(workingBranch, "wb-1")
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

	run := func(t *testing.T, r *singleBranchRemote, singleBranch bool, depth *int) outcome {
		dir := filepath.Join(t.TempDir(), "clone")
		g := &GoGit{}

		require.NoError(t, g.Clone("", "", r.url, dir, nil, depth, "main", singleBranch))
		require.NoError(t, g.Checkout("", "", "main", workingBranch, dir, true, depth))

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
		r.commit(workingBranch, "api-commit")

		require.NoError(t, g.Pull("", "", dir, workingBranch, singleBranch, true, depth))
		o.localAfterPull, err = g.GetLatestCommitHash(dir)
		require.NoError(t, err)
		assert.Equal(t, r.head(workingBranch), o.localAfterPull, "local branch should match the remote after Pull")
		return o
	}

	one := 1
	for _, depth := range []*int{nil, &one} {
		depthName := "full history"
		if depth != nil {
			depthName = "depth 1"
		}

		for _, tt := range tests {
			if depth != nil && tt.skipShallow {
				continue
			}

			t.Run(depthName+"/"+tt.name, func(t *testing.T) {
				// each mode gets its own identical remote
				rFull := newSingleBranchRemote(t)
				tt.setup(rFull)
				full := run(t, rFull, false, depth)

				rSingle := newSingleBranchRemote(t)
				tt.setup(rSingle)
				single := run(t, rSingle, true, depth)

				assert.Equal(t, full.forceReset, single.forceReset, "ForceReset should not depend on singleBranch")
				assert.Equal(t,
					full.localAfterCheckout == rFull.head("main"),
					single.localAfterCheckout == rSingle.head("main"),
					"working branch should start from the same base regardless of singleBranch")
			})
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
