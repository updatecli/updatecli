package bazelmod

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestTarget(t *testing.T) {
	tests := []struct {
		name            string
		setupContent    string
		spec            Spec
		source          string
		dryRun          bool
		expectedChanged bool
		wantErr         bool
		errorContains   string
		checkFile       func(t *testing.T, filePath string)
	}{
		{
			name: "Update version - dry run",
			setupContent: `module(name = "test_project", version = "1.0.0")

bazel_dep(name = "rules_go", version = "0.42.0")`,
			spec: Spec{
				File:   "MODULE.bazel",
				Module: "rules_go",
			},
			source:          "0.43.0",
			dryRun:          true,
			expectedChanged: true,
			wantErr:         false,
			checkFile: func(t *testing.T, filePath string) {
				// In dry run, file should not be changed
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)
				assert.Contains(t, string(content), `version = "0.42.0"`)
				assert.NotContains(t, string(content), `version = "0.43.0"`)
			},
		},
		{
			name: "Update version - actual update",
			setupContent: `module(name = "test_project", version = "1.0.0")

bazel_dep(name = "rules_go", version = "0.42.0")`,
			spec: Spec{
				File:   "MODULE.bazel",
				Module: "rules_go",
			},
			source:          "0.43.0",
			dryRun:          false,
			expectedChanged: true,
			wantErr:         false,
			checkFile: func(t *testing.T, filePath string) {
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)
				assert.Contains(t, string(content), `version = "0.43.0"`)
				assert.NotContains(t, string(content), `version = "0.42.0"`)
			},
		},
		{
			name: "Version already matches",
			setupContent: `module(name = "test_project", version = "1.0.0")

bazel_dep(name = "rules_go", version = "0.42.0")`,
			spec: Spec{
				File:   "MODULE.bazel",
				Module: "rules_go",
			},
			source:          "0.42.0",
			dryRun:          false,
			expectedChanged: false,
			wantErr:         false,
			checkFile: func(t *testing.T, filePath string) {
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)
				assert.Contains(t, string(content), `version = "0.42.0"`)
			},
		},
		{
			name: "Update multi-line bazel_dep",
			setupContent: `module(name = "test_project", version = "1.0.0")

bazel_dep(
    name = "protobuf",
    version = "21.7",
    repo_name = "com_google_protobuf",
)`,
			spec: Spec{
				File:   "MODULE.bazel",
				Module: "protobuf",
			},
			source:          "22.0",
			dryRun:          false,
			expectedChanged: true,
			wantErr:         false,
			checkFile: func(t *testing.T, filePath string) {
				content, err := os.ReadFile(filePath)
				require.NoError(t, err)
				assert.Contains(t, string(content), `version = "22.0"`)
				assert.NotContains(t, string(content), `version = "21.7"`)
				assert.Contains(t, string(content), `repo_name = "com_google_protobuf"`)
			},
		},
		{
			name: "Module not found",
			setupContent: `module(name = "test_project", version = "1.0.0")

bazel_dep(name = "rules_go", version = "0.42.0")`,
			spec: Spec{
				File:   "MODULE.bazel",
				Module: "nonexistent",
			},
			source:        "1.0.0",
			dryRun:        false,
			wantErr:       true,
			errorContains: "not found",
		},
		{
			name: "Empty source version",
			setupContent: `module(name = "test_project", version = "1.0.0")

bazel_dep(name = "rules_go", version = "0.42.0")`,
			spec: Spec{
				File:   "MODULE.bazel",
				Module: "rules_go",
			},
			source:        "",
			dryRun:        false,
			wantErr:       true,
			errorContains: "no version provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for test files
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "MODULE.bazel")
			err := os.WriteFile(filePath, []byte(tt.setupContent), 0600)
			require.NoError(t, err)

			// Update spec to use absolute path
			tt.spec.File = filePath
			b, err := New(tt.spec)
			require.NoError(t, err)

			resultTarget := &result.Target{}
			err = b.Target(tt.source, nil, tt.dryRun, resultTarget)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedChanged, resultTarget.Changed)
			if tt.checkFile != nil {
				tt.checkFile(t, filePath)
			}
		})
	}
}
