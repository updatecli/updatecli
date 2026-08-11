package gitcommit

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func TestCondition(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		spec        Spec
		scm         scm.ScmHandler
		handler     *mockGitHandler
		wantPass    bool
		wantMessage string
		wantDir     string
		wantCommit  string
		wantErr     string
	}{
		{
			name:        "source commit found in SCM working directory",
			source:      "abc123",
			scm:         &scm.MockScm{WorkingDir: "/tmp/scm"},
			handler:     &mockGitHandler{exists: true},
			wantPass:    true,
			wantMessage: `git commit "abc123" found`,
			wantDir:     "/tmp/scm",
			wantCommit:  "abc123",
		},
		{
			name:        "spec.hash overrides the source input",
			source:      "abc123",
			spec:        Spec{Path: "/tmp/repository", Hash: "def456"},
			handler:     &mockGitHandler{exists: true},
			wantPass:    true,
			wantMessage: `git commit "def456" found`,
			wantDir:     "/tmp/repository",
			wantCommit:  "def456",
		},
		{
			name:        "commit not found",
			source:      "abc123",
			spec:        Spec{Path: "/tmp/repository"},
			handler:     &mockGitHandler{},
			wantMessage: `git commit "abc123" not found`,
			wantDir:     "/tmp/repository",
			wantCommit:  "abc123",
		},
		{
			name:    "missing working directory",
			source:  "abc123",
			handler: &mockGitHandler{},
			wantErr: "unknown Git working directory",
		},
		{
			name:    "missing commit",
			scm:     &scm.MockScm{WorkingDir: "/tmp/scm"},
			handler: &mockGitHandler{},
			wantErr: "unknown Git commit",
		},
		{
			name:       "Git lookup error",
			source:     "abc123",
			spec:       Spec{Path: "/tmp/repository"},
			handler:    &mockGitHandler{err: errors.New("corrupted repository")},
			wantDir:    "/tmp/repository",
			wantCommit: "abc123",
			wantErr:    "checking Git commit existence: corrupted repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource := &GitCommit{spec: tt.spec, nativeGitHandler: tt.handler}
			pass, message, err := resource.Condition(context.Background(), tt.source, tt.scm)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.False(t, pass)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantPass, pass)
				assert.Equal(t, tt.wantMessage, message)
			}
			assert.Equal(t, tt.wantDir, tt.handler.gotDirectory)
			assert.Equal(t, tt.wantCommit, tt.handler.gotCommit)
		})
	}
}
