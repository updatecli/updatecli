package bazelmod

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestSource(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Copy test files to temp directory
	testFiles := map[string]string{
		"MODULE.bazel": `
		module(name = "test_project", version = "1.0.0")

		bazel_dep(name = "rules_go", version = "0.42.0")
		bazel_dep(name = "gazelle", version = "0.34.0")
		bazel_dep(
			name = "protobuf",
			version = "21.7",
			repo_name = "com_google_protobuf",
		)`,
		"MODULE_simple.bazel": `module(name = "simple_project", version = "1.0.0")

		bazel_dep(name = "rules_go", version = "0.42.0")`,
	}

	for filename, content := range testFiles {
		filePath := filepath.Join(tmpDir, filename)
		err := os.WriteFile(filePath, []byte(content), 0644)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		spec          Spec
		workingDir    string
		expectedValue string
		wantErr       bool
		errorContains string
	}{
		{
			name: "Read existing module version",
			spec: Spec{
				File:   "MODULE.bazel",
				Module: "rules_go",
			},
			workingDir:    tmpDir,
			expectedValue: "0.42.0",
			wantErr:       false,
		},
		{
			name: "Read multi-line module version",
			spec: Spec{
				File:   "MODULE.bazel",
				Module: "protobuf",
			},
			workingDir:    tmpDir,
			expectedValue: "21.7",
			wantErr:       false,
		},
		{
			name: "Read from simple file",
			spec: Spec{
				File:   "MODULE_simple.bazel",
				Module: "rules_go",
			},
			workingDir:    tmpDir,
			expectedValue: "0.42.0",
			wantErr:       false,
		},
		{
			name: "Module not found",
			spec: Spec{
				File:   "MODULE.bazel",
				Module: "nonexistent",
			},
			workingDir:    tmpDir,
			wantErr:       true,
			errorContains: "not found",
		},
		{
			name: "File not found",
			spec: Spec{
				File:   "nonexistent.bazel",
				Module: "rules_go",
			},
			workingDir:    tmpDir,
			wantErr:       true,
			errorContains: "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := New(tt.spec)
			require.NoError(t, err)

			resultSource := &result.Source{}
			err = b.Source(tt.workingDir, resultSource)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, result.SUCCESS, resultSource.Result)
			assert.Equal(t, tt.expectedValue, resultSource.Information)
		})
	}
}
