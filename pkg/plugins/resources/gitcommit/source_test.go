package gitcommit

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
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

// processWorkingDirectory is where a git resource looks for a repository when the
// manifest names neither an scm, a url nor a path.
func processWorkingDirectory() string {
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "."
	}
	return workingDirectory
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
			// Without an scm, a url or a spec.path, the repository is the one holding
			// the process working directory: that is what "a relative path resolves
			// from where updatecli was started" means for a git resource.
			name:     "no working directory falls back to the process directory",
			handler:  &mockGitHandler{hash: "abc123"},
			wantHash: "abc123",
			wantDir:  processWorkingDirectory(),
			// spec.branch is empty, "HEAD" is only how the description spells it.
			wantBranch: "",
			wantDesc:   `Git commit "abc123" found for branch "HEAD"`,
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
			err := resource.Source(context.Background(), utils.Resolver{BaseDir: tt.workingDir, Boundary: tt.workingDir}, &got)
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
