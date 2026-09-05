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

func TestIsCommitExist(t *testing.T) {
	directory := t.TempDir()
	repository, err := git.PlainInit(directory, false)
	require.NoError(t, err)

	worktree, err := repository.Worktree()
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(directory+"/file.txt", []byte("content"), 0o600))
	_, err = worktree.Add("file.txt")
	require.NoError(t, err)
	hash, err := worktree.Commit("initial commit", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Updatecli Test",
			Email: "test@updatecli.io",
			When:  time.Now(),
		},
	})
	require.NoError(t, err)

	handler := GoGit{}
	tests := []struct {
		name       string
		commit     string
		want       bool
		wantErr    string
		workingDir string
	}{
		{name: "full hash", commit: hash.String(), want: true, workingDir: directory},
		{name: "abbreviated hash", commit: hash.String()[:7], want: true, workingDir: directory},
		{name: "missing commit", commit: "0123456789012345678901234567890123456789", workingDir: directory},
		{name: "unknown revision", commit: "not-a-commit", workingDir: directory},
		{name: "invalid repository", commit: hash.String(), wantErr: "opening", workingDir: directory + "/missing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := handler.IsCommitExist(tt.workingDir, tt.commit)
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

// commitDaysAgo builds the committer date of a commit created n days ago.
func commitDaysAgo(n int) time.Time {
	return time.Now().Add(-time.Duration(n) * 24 * time.Hour)
}

// olderThan builds a matcher accepting the commits which are at least n days old.
func olderThan(n int) func(time.Time) bool {
	return func(when time.Time) bool {
		return when.Before(commitDaysAgo(n))
	}
}

func TestSearchCommit(t *testing.T) {
	directory := t.TempDir()
	repository, err := git.PlainInit(directory, false)
	require.NoError(t, err)

	worktree, err := repository.Worktree()
	require.NoError(t, err)

	commit := func(content, message string, when time.Time) plumbing.Hash {
		require.NoError(t, os.WriteFile(directory+"/file.txt", []byte(content), 0o600))
		_, err = worktree.Add("file.txt")
		require.NoError(t, err)
		hash, commitErr := worktree.Commit(message, &git.CommitOptions{
			AllowEmptyCommits: true,
			Author: &object.Signature{
				Name:  "Updatecli Test",
				Email: "test@updatecli.io",
				When:  when,
			},
		})
		require.NoError(t, commitErr)
		return hash
	}

	/*
		A history with a merge, so that the walk order is actually exercised:

		  master  30d --- 10d ------------- merge (0d)
		                     \             /
		  feature             `--- 5d ----'

		A depth-first walk reaches the 5 days old feature commit before the 10 days old
		master one, so a matcher accepting anything older than 3 days only returns the
		right commit when the history is walked by committer time.
	*/
	oldest := commit("oldest", "oldest commit", commitDaysAgo(30))
	mainline := commit("mainline", "mainline commit", commitDaysAgo(10))

	featureReference := plumbing.NewHashReference(plumbing.NewBranchReferenceName("feature"), mainline)
	require.NoError(t, repository.Storer.SetReference(featureReference))
	require.NoError(t, worktree.Checkout(&git.CheckoutOptions{Branch: featureReference.Name()}))
	feature := commit("feature", "feature commit", commitDaysAgo(5))

	require.NoError(t, worktree.Checkout(&git.CheckoutOptions{Branch: plumbing.NewBranchReferenceName("master")}))
	merge := commit("merge", "merge commit", commitDaysAgo(0))
	mergeCommit, err := repository.CommitObject(merge)
	require.NoError(t, err)
	mergeCommit.ParentHashes = []plumbing.Hash{mainline, feature}
	mergeObject := repository.Storer.NewEncodedObject()
	require.NoError(t, mergeCommit.Encode(mergeObject))
	merge, err = repository.Storer.SetEncodedObject(mergeObject)
	require.NoError(t, err)
	require.NoError(t, repository.Storer.SetReference(
		plumbing.NewHashReference(plumbing.NewBranchReferenceName("master"), merge)))
	require.NoError(t, worktree.Reset(&git.ResetOptions{Commit: merge, Mode: git.HardReset}))

	remoteReference := plumbing.NewHashReference(
		plumbing.NewRemoteReferenceName(DefaultRemoteReferenceName, "remote-only"),
		mainline,
	)
	require.NoError(t, repository.Storer.SetReference(remoteReference))

	handler := GoGit{}

	t.Run("a matching tip stops the walk immediately", func(t *testing.T) {
		visited := 0
		got, err := handler.SearchCommit(directory, "", func(time.Time) bool {
			visited++
			return true
		})
		require.NoError(t, err)
		assert.Equal(t, merge.String(), got.Hash)
		// The laziness of the walk is the point: a matching tip must cost one commit.
		assert.Equal(t, 1, visited)
	})

	t.Run("the newest matching commit wins across a merge", func(t *testing.T) {
		got, err := handler.SearchCommit(directory, "", olderThan(3))
		require.NoError(t, err)
		assert.Equal(t, feature.String(), got.Hash)
	})

	t.Run("the history is walked by committer time, not depth first", func(t *testing.T) {
		got, err := handler.SearchCommit(directory, "", olderThan(7))
		require.NoError(t, err)
		assert.Equal(t, mainline.String(), got.Hash)
	})

	t.Run("the oldest commit is reachable", func(t *testing.T) {
		got, err := handler.SearchCommit(directory, "", olderThan(20))
		require.NoError(t, err)
		assert.Equal(t, oldest.String(), got.Hash)
		// The committer date is what the matcher was given.
		assert.WithinDuration(t, commitDaysAgo(30), got.When, time.Minute)
	})

	t.Run("an exhausted history reports no commit found", func(t *testing.T) {
		_, err := handler.SearchCommit(directory, "", olderThan(365))
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNoCommitFound)
	})

	t.Run("a local branch is resolved", func(t *testing.T) {
		got, err := handler.SearchCommit(directory, "feature", func(time.Time) bool { return true })
		require.NoError(t, err)
		assert.Equal(t, feature.String(), got.Hash)
	})

	t.Run("a remote branch is resolved", func(t *testing.T) {
		got, err := handler.SearchCommit(directory, "remote-only", func(time.Time) bool { return true })
		require.NoError(t, err)
		assert.Equal(t, mainline.String(), got.Hash)
	})

	t.Run("a missing branch errors", func(t *testing.T) {
		_, err := handler.SearchCommit(directory, "missing", func(time.Time) bool { return true })
		require.Error(t, err)
		assert.Contains(t, err.Error(), `branch "missing" not found`)
	})

	t.Run("an invalid repository errors", func(t *testing.T) {
		_, err := handler.SearchCommit(directory+"/missing", "", func(time.Time) bool { return true })
		require.Error(t, err)
		assert.Contains(t, err.Error(), "opening")
	})
}

// TestSearchCommitTruncatedHistory covers the boundary of a shallow repository, whose
// oldest commit references parents which were never fetched.
func TestSearchCommitTruncatedHistory(t *testing.T) {
	directory := t.TempDir()
	repository, err := git.PlainInit(directory, false)
	require.NoError(t, err)

	worktree, err := repository.Worktree()
	require.NoError(t, err)

	commit := func(content, message string, when time.Time) plumbing.Hash {
		require.NoError(t, os.WriteFile(directory+"/file.txt", []byte(content), 0o600))
		_, err = worktree.Add("file.txt")
		require.NoError(t, err)
		hash, commitErr := worktree.Commit(message, &git.CommitOptions{
			Author: &object.Signature{
				Name:  "Updatecli Test",
				Email: "test@updatecli.io",
				When:  when,
			},
		})
		require.NoError(t, commitErr)
		return hash
	}

	boundaryHash := commit("boundary", "boundary commit", commitDaysAgo(10))
	tip := commit("tip", "tip commit", commitDaysAgo(1))

	// Point the oldest commit at a parent absent from the object store, exactly as the
	// oldest commit of a shallow clone does.
	boundary, err := repository.CommitObject(boundaryHash)
	require.NoError(t, err)
	boundary.ParentHashes = []plumbing.Hash{plumbing.NewHash("0123456789abcdef0123456789abcdef01234567")}

	encoded := repository.Storer.NewEncodedObject()
	require.NoError(t, boundary.Encode(encoded))
	truncatedBoundary, err := repository.Storer.SetEncodedObject(encoded)
	require.NoError(t, err)

	// Re-parent the tip onto the rewritten boundary commit.
	tipCommit, err := repository.CommitObject(tip)
	require.NoError(t, err)
	tipCommit.ParentHashes = []plumbing.Hash{truncatedBoundary}
	encodedTip := repository.Storer.NewEncodedObject()
	require.NoError(t, tipCommit.Encode(encodedTip))
	truncatedTip, err := repository.Storer.SetEncodedObject(encodedTip)
	require.NoError(t, err)
	require.NoError(t, repository.Storer.SetReference(
		plumbing.NewHashReference(plumbing.NewBranchReferenceName("master"), truncatedTip)))

	handler := GoGit{}

	/*
		The commit on the shallow boundary must be matched before its missing parents are
		looked up. The go-git iterators load the parents of a commit before handing it
		over, so they drop this commit entirely -- which is the one a `depth` limited
		clone is the most likely to want.
	*/
	t.Run("the commit on the boundary is still matched", func(t *testing.T) {
		got, err := handler.SearchCommit(directory, "", olderThan(5))
		require.NoError(t, err)
		assert.Equal(t, truncatedBoundary.String(), got.Hash)
	})

	t.Run("a newer commit is preferred over the boundary", func(t *testing.T) {
		got, err := handler.SearchCommit(directory, "", func(time.Time) bool { return true })
		require.NoError(t, err)
		assert.Equal(t, truncatedTip.String(), got.Hash)
	})

	t.Run("walking past the boundary reports a truncated history rather than failing", func(t *testing.T) {
		_, err := handler.SearchCommit(directory, "", olderThan(365))
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNoCommitFound)
		assert.ErrorIs(t, err, ErrTruncatedHistory)
	})
}
