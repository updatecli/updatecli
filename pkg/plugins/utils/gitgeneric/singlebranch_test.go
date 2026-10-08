package gitgeneric

import (
	"os"
	"os/exec"
	"path/filepath"
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
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
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
	return trimNL(gitCmd(r.t, r.seed, "rev-parse", branch))
}

func trimNL(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
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
		},
	}

	type outcome struct {
		localAfterCheckout string
		forceReset         bool
		localAfterPull     string
	}

	run := func(t *testing.T, r *singleBranchRemote, singleBranch bool) outcome {
		dir := filepath.Join(t.TempDir(), "clone")
		g := &GoGit{}

		require.NoError(t, g.Clone("", "", r.url, dir, nil, nil, "main", singleBranch))
		require.NoError(t, g.Checkout("", "", "main", workingBranch, dir, true, nil))

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

		require.NoError(t, g.Pull("", "", dir, workingBranch, singleBranch, true, nil))
		o.localAfterPull, err = g.GetLatestCommitHash(dir)
		require.NoError(t, err)
		assert.Equal(t, r.head(workingBranch), o.localAfterPull, "local branch should match the remote after Pull")
		return o
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// each mode gets its own identical remote
			rFull := newSingleBranchRemote(t)
			tt.setup(rFull)
			full := run(t, rFull, false)

			rSingle := newSingleBranchRemote(t)
			tt.setup(rSingle)
			single := run(t, rSingle, true)

			assert.Equal(t, full.forceReset, single.forceReset, "ForceReset should not depend on singleBranch")
			assert.Equal(t,
				full.localAfterCheckout == rFull.head("main"),
				single.localAfterCheckout == rSingle.head("main"),
				"working branch should start from the same base regardless of singleBranch")
		})
	}
}
