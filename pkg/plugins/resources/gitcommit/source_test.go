package gitcommit

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/gitgeneric"
)

type mockGitHandler struct {
	hash   string
	exists bool
	err    error
	// commits holds the branch history, newest first.
	commits      []gitgeneric.DatedCommit
	gotDirectory string
	gotBranch    string
	gotCommit    string
	gotSearch    bool
}

// SearchCommit reproduces the newest-first, stop-on-first-match contract of the real
// handler so that the age filtering itself is exercised rather than a canned answer.
func (m *mockGitHandler) SearchCommit(workingDir, branch string, match func(when time.Time) bool) (gitgeneric.DatedCommit, error) {
	m.gotDirectory = workingDir
	m.gotBranch = branch
	m.gotSearch = true

	if m.err != nil {
		return gitgeneric.DatedCommit{}, m.err
	}

	for _, commit := range m.commits {
		if match(commit.When) {
			return commit, nil
		}
	}

	return gitgeneric.DatedCommit{}, fmt.Errorf("%w in the history of %q", gitgeneric.ErrNoCommitFound, branch)
}

// commitDaysAgo builds the committer date of a commit created n days ago.
func commitDaysAgo(n int) time.Time {
	return time.Now().Add(-time.Duration(n) * 24 * time.Hour)
}

func (m *mockGitHandler) GetCommitHash(workingDir, branch string) (string, error) {
	m.gotDirectory = workingDir
	m.gotBranch = branch
	return m.hash, m.err
}

func (m *mockGitHandler) IsCommitExist(workingDir, commit string) (bool, error) {
	m.gotDirectory = workingDir
	m.gotCommit = commit
	return m.exists, m.err
}

func TestSource(t *testing.T) {
	// A branch history, newest first, as the mocked handler walks it.
	commits := []gitgeneric.DatedCommit{
		{Hash: "ghi789", When: commitDaysAgo(1)},
		{Hash: "def456", When: commitDaysAgo(10)},
		{Hash: "abc123", When: commitDaysAgo(30)},
	}

	tests := []struct {
		name        string
		workingDir  string
		spec        Spec
		handler     *mockGitHandler
		wantHash    string
		wantDir     string
		wantBranch  string
		wantDesc    string
		wantSkipped bool
		wantSearch  bool
		wantErr     string
	}{
		{
			name:       "SCM working directory HEAD",
			workingDir: "/tmp/scm",
			handler:    &mockGitHandler{hash: "abc123"},
			wantHash:   "abc123",
			wantDir:    "/tmp/scm",
			wantDesc:   `Git commit "abc123" found for branch "HEAD"`,
		},
		{
			name:       "configured path and branch",
			workingDir: "/tmp/scm",
			spec:       Spec{Path: "/tmp/repository", Branch: "release"},
			handler:    &mockGitHandler{hash: "def456"},
			wantHash:   "def456",
			wantDir:    "/tmp/repository",
			wantBranch: "release",
			wantDesc:   `Git commit "def456" found for branch "release"`,
		},
		{
			name:    "missing working directory",
			handler: &mockGitHandler{},
			wantErr: "unknown Git working directory",
		},
		{
			name:       "Git lookup error",
			workingDir: "/tmp/scm",
			spec:       Spec{Branch: "missing"},
			handler:    &mockGitHandler{err: errors.New("branch not found")},
			wantDir:    "/tmp/scm",
			wantBranch: "missing",
			wantErr:    "retrieving latest Git commit: branch not found",
		},
		{
			name:       "the branch tip is still in cooldown",
			workingDir: "/tmp/scm",
			spec:       Spec{Age: age.Spec{Minimum: "7d"}},
			handler:    &mockGitHandler{commits: commits},
			wantHash:   "def456",
			wantDir:    "/tmp/scm",
			wantSearch: true,
		},
		{
			name:       "commits older than the maximum are discarded",
			workingDir: "/tmp/scm",
			spec:       Spec{Age: age.Spec{Maximum: "20d"}},
			handler:    &mockGitHandler{commits: commits},
			wantHash:   "ghi789",
			wantDir:    "/tmp/scm",
			wantSearch: true,
		},
		{
			name:        "the source is skipped while every commit is still in cooldown",
			workingDir:  "/tmp/scm",
			spec:        Spec{Age: age.Spec{Minimum: "60d"}},
			handler:     &mockGitHandler{commits: commits},
			wantDir:     "/tmp/scm",
			wantSkipped: true,
			wantSearch:  true,
		},
		{
			name:        "the source is skipped when the age window falls between two commits",
			workingDir:  "/tmp/scm",
			spec:        Spec{Age: age.Spec{Minimum: "12d", Maximum: "20d"}},
			handler:     &mockGitHandler{commits: commits},
			wantDir:     "/tmp/scm",
			wantSkipped: true,
			wantSearch:  true,
		},
		{
			name:       "a Git lookup error is not mistaken for a running cooldown",
			workingDir: "/tmp/scm",
			spec:       Spec{Branch: "missing", Age: age.Spec{Minimum: "7d"}},
			handler:    &mockGitHandler{err: errors.New("branch not found")},
			wantDir:    "/tmp/scm",
			wantBranch: "missing",
			wantSearch: true,
			wantErr:    "retrieving latest Git commit: branch not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := &GitCommit{spec: tt.spec, nativeGitHandler: tt.handler}
			got := result.Source{}
			err := resource.Source(context.Background(), tt.workingDir, &got)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)

				switch tt.wantSkipped {
				case true:
					assert.Equal(t, result.SKIPPED, got.Result)
					assert.Equal(t, "no git commit matches the age filter yet", got.Description)
					assert.Empty(t, got.Information)
				case false:
					assert.Equal(t, result.SUCCESS, got.Result)
					assert.Equal(t, tt.wantHash, got.Information)
					if tt.wantDesc != "" {
						assert.Equal(t, tt.wantDesc, got.Description)
					}
				}
			}
			assert.Equal(t, tt.wantDir, tt.handler.gotDirectory)
			assert.Equal(t, tt.wantBranch, tt.handler.gotBranch)
			// Without an age filter the history must never be walked.
			assert.Equal(t, tt.wantSearch, tt.handler.gotSearch)
		})
	}
}
