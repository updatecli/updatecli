package gitgeneric

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// baseBranch is a branch, other than the default branch of the remote, used as the base branch.
const baseBranch = "9.6"

// newNonDefaultBaseRemote creates a remote whose default branch is main, with a baseBranch
// branch diverging from it.
func newNonDefaultBaseRemote(t *testing.T) *singleBranchRemote {
	t.Helper()
	r := newSingleBranchRemote(t)
	r.commit(baseBranch, "base-1")
	r.commit("main", "main-2")
	return r
}

// TestSingleBranchNonDefaultBaseBranch checks that Checkout works with a single branch clone
// of a branch that isn't the default branch of the remote, whose commits aren't fetched,
// and gives the same result as with a regular clone.
func TestSingleBranchNonDefaultBaseBranch(t *testing.T) {
	tests := []struct {
		name string
		// setup prepares the remote before updatecli runs.
		setup func(r *singleBranchRemote, branch string)
		// existsRemotely tells whether the working branch exists on the remote.
		existsRemotely bool
	}{
		{
			name:  "working branch does not exist yet",
			setup: func(r *singleBranchRemote, branch string) {},
		},
		{
			name: "working branch is ahead of the base branch",
			setup: func(r *singleBranchRemote, branch string) {
				gitCmd(t, r.seed, "checkout", "-q", baseBranch)
				r.commit(branch, "wb-1")
			},
			existsRemotely: true,
		},
	}

	type outcome struct {
		localAfterCheckout string
		forceReset         bool
	}

	// run checks out the working branch twice, as it happens when a pipeline has several targets.
	run := func(t *testing.T, r *singleBranchRemote, branch string, existsRemotely, singleBranch bool, depth *int) outcome {
		dir := filepath.Join(t.TempDir(), "clone")
		g := &GoGit{}

		require.NoError(t, g.Clone("", "", r.url, dir, nil, depth, baseBranch, singleBranch))
		refSpecs := remoteFetchRefSpecs(t, dir)

		expected := r.head(baseBranch)
		if existsRemotely {
			expected = r.head(branch)
		}

		var o outcome
		for i := range 2 {
			require.NoError(t, g.Checkout("", "", baseBranch, branch, dir, true, depth), "checkout %d", i+1)

			hash, err := g.GetLatestCommitHash(dir)
			require.NoError(t, err)
			assert.Equal(t, expected, hash, "unexpected working branch commit after checkout %d", i+1)

			if i == 0 {
				o = outcome{localAfterCheckout: hash, forceReset: g.IsForceReset()}
			}
		}

		assert.Equal(t, refSpecs, remoteFetchRefSpecs(t, dir), "remote configuration should be restored after checkout")
		return o
	}

	one := 1
	for _, tt := range tests {
		t.Run("full history/"+tt.name, func(t *testing.T) {
			rFull := newNonDefaultBaseRemote(t)
			tt.setup(rFull, workingBranch)
			full := run(t, rFull, workingBranch, tt.existsRemotely, false, nil)

			rSingle := newNonDefaultBaseRemote(t)
			tt.setup(rSingle, workingBranch)
			single := run(t, rSingle, workingBranch, tt.existsRemotely, true, nil)

			assert.Equal(t, full.forceReset, single.forceReset, "ForceReset should not depend on singleBranch")
		})

		// A regular depth 1 clone isn't compared: its pull resolves the remote HEAD, whose shallow
		// history doesn't always allow go-git to tell whether it's a fast-forward.
		t.Run("depth 1/"+tt.name, func(t *testing.T) {
			r := newNonDefaultBaseRemote(t)
			tt.setup(r, workingBranch)
			run(t, r, workingBranch, tt.existsRemotely, true, &one)
		})
	}
}

// TestSingleBranchNonDefaultBaseBranchIsWorkingBranch checks the case where no separate working
// branch is used, with a single branch clone of a branch that isn't the default branch of the remote.
func TestSingleBranchNonDefaultBaseBranchIsWorkingBranch(t *testing.T) {
	r := newNonDefaultBaseRemote(t)
	dir := filepath.Join(t.TempDir(), "clone")
	g := &GoGit{}

	require.NoError(t, g.Clone("", "", r.url, dir, nil, nil, baseBranch, true))

	r.commit(baseBranch, "base-2")

	require.NoError(t, g.Checkout("", "", baseBranch, baseBranch, dir, true, nil))
	assert.False(t, g.IsForceReset())

	hash, err := g.GetLatestCommitHash(dir)
	require.NoError(t, err)
	assert.Equal(t, r.head(baseBranch), hash, "the base branch should be pulled")
}

// TestSingleBranchNonDefaultBaseBranchExistingClone checks that cloning again into a directory
// holding a single branch clone of a branch that isn't the default branch of the remote,
// such as one left by a previous execution, pulls the configured branch.
func TestSingleBranchNonDefaultBaseBranchExistingClone(t *testing.T) {
	r := newNonDefaultBaseRemote(t)
	dir := filepath.Join(t.TempDir(), "clone")
	g := &GoGit{}

	require.NoError(t, g.Clone("", "", r.url, dir, nil, nil, baseBranch, true))

	r.commit(baseBranch, "base-2")

	require.NoError(t, g.Clone("", "", r.url, dir, nil, nil, baseBranch, true))

	hash, err := g.GetLatestCommitHash(dir)
	require.NoError(t, err)
	assert.Equal(t, r.head(baseBranch), hash, "the configured branch should be pulled")
}

// TestSingleBranchNonDefaultBaseBranchLocalOnlyWorkingBranch checks that a working branch which
// only exists locally, and was never pushed, is neither lost nor failing when it's checked out again.
// Its pull has no matching reference on the remote, which is not an error.
func TestSingleBranchNonDefaultBaseBranchLocalOnlyWorkingBranch(t *testing.T) {
	r := newNonDefaultBaseRemote(t)
	dir := filepath.Join(t.TempDir(), "clone")
	g := &GoGit{}

	require.NoError(t, g.Clone("", "", r.url, dir, nil, nil, baseBranch, true))
	require.NoError(t, g.Checkout("", "", baseBranch, workingBranch, dir, true, nil))

	// A local commit, not pushed, as done by updatecli before it pushes a working branch.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "local"), []byte("local"), 0o600))
	gitCmd(t, dir, "add", "local")
	gitCmd(t, dir, "commit", "-qm", "local")
	local := strings.TrimSpace(gitCmd(t, dir, "rev-parse", "HEAD"))

	refSpecs := remoteFetchRefSpecs(t, dir)

	require.NoError(t, g.Checkout("", "", baseBranch, workingBranch, dir, true, nil))

	hash, err := g.GetLatestCommitHash(dir)
	require.NoError(t, err)
	assert.Equal(t, local, hash, "the local commit of the working branch should be kept")
	assert.Equal(t, refSpecs, remoteFetchRefSpecs(t, dir), "remote configuration should be restored after checkout")
}

// TestSingleBranchNonDefaultBaseBranchPull checks that Pull, used after a commit is created
// remotely with the GitHub API, works with a single branch clone of a branch that isn't the
// default branch of the remote.
func TestSingleBranchNonDefaultBaseBranchPull(t *testing.T) {
	one := 1

	tests := []struct {
		name  string
		depth *int
	}{
		{name: "full history"},
		{name: "depth 1", depth: &one},
	}

	for _, tt := range tests {
		for _, existsRemotely := range []bool{false, true} {
			name := tt.name + "/working branch does not exist yet"
			if existsRemotely {
				name = tt.name + "/working branch already exists"
			}

			t.Run(name, func(t *testing.T) {
				r := newNonDefaultBaseRemote(t)
				if existsRemotely {
					gitCmd(t, r.seed, "checkout", "-q", baseBranch)
					r.commit(workingBranch, "wb-1")
				}

				dir := filepath.Join(t.TempDir(), "clone")
				g := &GoGit{}

				require.NoError(t, g.Clone("", "", r.url, dir, nil, tt.depth, baseBranch, true))
				require.NoError(t, g.Checkout("", "", baseBranch, workingBranch, dir, true, tt.depth))

				// Another actor, the GitHub API, creates a commit on the working branch.
				gitCmd(t, r.seed, "checkout", "-q", baseBranch)
				r.commit(workingBranch, "wb-api")

				refSpecs := remoteFetchRefSpecs(t, dir)

				require.NoError(t, g.Pull("", "", dir, workingBranch, true, true, tt.depth))

				hash, err := g.GetLatestCommitHash(dir)
				require.NoError(t, err)
				assert.Equal(t, r.head(workingBranch), hash, "the commit created remotely should be pulled")
				assert.Equal(t, refSpecs, remoteFetchRefSpecs(t, dir), "remote configuration should be restored after pull")
			})
		}
	}
}
