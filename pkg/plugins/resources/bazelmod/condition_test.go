package bazelmod

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCondition(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Create test file
	filePath := filepath.Join(tmpDir, "MODULE.bazel")
	content := `
	module(name = "test_project", version = "1.0.0")

	bazel_dep(name = "rules_go", version = "0.42.0")
	bazel_dep(name = "gazelle", version = "0.34.0")
	bazel_dep(
		name = "protobuf",
		version = "21.7",
		repo_name = "com_google_protobuf",
	)`
	err := os.WriteFile(filePath, []byte(content), 0644)
	require.NoError(t, err)

	tests := []struct {
		name            string
		spec            Spec
		source          string
		rootDir         string
		expectedPass    bool
		wantErr         bool
		errorContains   string
		messageContains string
	}{
		{
			name: "Version matches",
			spec: Spec{
				File:   filepath.Join(tmpDir, "MODULE.bazel"),
				Module: "rules_go",
			},
			source:          "0.42.0",
			rootDir:         "",
			expectedPass:    true,
			wantErr:         false,
			messageContains: "correctly set",
		},
		{
			name: "Version does not match",
			spec: Spec{
				File:   filepath.Join(tmpDir, "MODULE.bazel"),
				Module: "rules_go",
			},
			source:          "0.43.0",
			rootDir:         "",
			expectedPass:    false,
			wantErr:         false,
			messageContains: "incorrectly set",
		},
		{
			name: "Multi-line module version matches",
			spec: Spec{
				File:   filepath.Join(tmpDir, "MODULE.bazel"),
				Module: "protobuf",
			},
			source:          "21.7",
			rootDir:         "",
			expectedPass:    true,
			wantErr:         false,
			messageContains: "correctly set",
		},
		{
			name: "Module not found",
			spec: Spec{
				File:   filepath.Join(tmpDir, "MODULE.bazel"),
				Module: "nonexistent",
			},
			source:        "1.0.0",
			rootDir:       "",
			wantErr:       true,
			errorContains: "not found",
		},
		{
			name: "File not found",
			spec: Spec{
				File:   filepath.Join(tmpDir, "nonexistent.bazel"),
				Module: "rules_go",
			},
			source:        "0.42.0",
			rootDir:       "",
			wantErr:       true,
			errorContains: "does not exist",
		},
		{
			name: "Empty source version",
			spec: Spec{
				File:   filepath.Join(tmpDir, "MODULE.bazel"),
				Module: "rules_go",
			},
			source:        "",
			rootDir:       "",
			wantErr:       true,
			errorContains: "no version provided",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := New(tt.spec)
			require.NoError(t, err)

			pass, message, err := b.Condition(tt.source, nil)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedPass, pass)
			if tt.messageContains != "" {
				assert.Contains(t, message, tt.messageContains)
			}
		})
	}
}
