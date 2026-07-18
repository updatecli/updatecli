package gitcommit

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

type mockGitHandler struct {
	hash         string
	exists       bool
	err          error
	gotDirectory string
	gotBranch    string
	gotCommit    string
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
	tests := []struct {
		name       string
		workingDir string
		spec       Spec
		handler    *mockGitHandler
		wantHash   string
		wantDir    string
		wantBranch string
		wantDesc   string
		wantErr    string
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
				assert.Equal(t, result.SUCCESS, got.Result)
				assert.Equal(t, tt.wantHash, got.Information)
				assert.Equal(t, tt.wantDesc, got.Description)
			}
			assert.Equal(t, tt.wantDir, tt.handler.gotDirectory)
			assert.Equal(t, tt.wantBranch, tt.handler.gotBranch)
		})
	}
}
