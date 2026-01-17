package bazel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseModuleDependencies(t *testing.T) {
	testdata := []struct {
		name         string
		moduleFile   string
		expectedDeps []dependency
		expectError  bool
	}{
		{
			name:       "Parse valid MODULE.bazel with multiple dependencies",
			moduleFile: "testdata/project1/MODULE.bazel",
			expectedDeps: []dependency{
				{Name: "rules_go", Version: "0.42.0"},
				{Name: "gazelle", Version: "0.34.0"},
				{Name: "protobuf", Version: "21.7"},
			},
			expectError: false,
		},
		{
			name:       "Parse valid MODULE.bazel with single dependency",
			moduleFile: "testdata/project2/subdir/MODULE.bazel",
			expectedDeps: []dependency{
				{Name: "rules_docker", Version: "0.26.0"},
			},
			expectError: false,
		},
		{
			name:        "Parse invalid MODULE.bazel (parser is lenient and may still extract deps)",
			moduleFile:  "testdata/invalid/MODULE.bazel",
			expectedDeps: []dependency{
				{Name: "rules_go", Version: "0.42.0"}, // Parser extracts this despite missing comma
			},
			expectError: false, // Parser is lenient and may not error on syntax issues
		},
		{
			name:        "Non-existent file",
			moduleFile:  "testdata/nonexistent/MODULE.bazel",
			expectedDeps: nil,
			expectError:  true,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			// Get absolute path
			absPath, err := filepath.Abs(tt.moduleFile)
			if err != nil {
				absPath = tt.moduleFile
			}

			deps, err := parseModuleDependencies(absPath)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, deps, len(tt.expectedDeps))

			// Check that all expected dependencies are present
			for _, expected := range tt.expectedDeps {
				found := false
				for _, dep := range deps {
					if dep.Name == expected.Name && dep.Version == expected.Version {
						found = true
						assert.Equal(t, absPath, dep.File)
						break
					}
				}
				assert.True(t, found, "Expected dependency %q not found", expected.Name)
			}
		})
	}
}

func TestParseModuleDependenciesEmptyFile(t *testing.T) {
	// Create a temporary file with only module declaration (no dependencies)
	tmpDir := t.TempDir()
	moduleFile := filepath.Join(tmpDir, "MODULE.bazel")
	content := `module(
    name = "test",
    version = "1.0.0",
)
`
	err := os.WriteFile(moduleFile, []byte(content), 0644)
	require.NoError(t, err)

	deps, err := parseModuleDependencies(moduleFile)
	require.NoError(t, err)
	assert.Len(t, deps, 0)
}

func TestParseModuleDependenciesMultiLine(t *testing.T) {
	// Create a temporary file with multi-line bazel_dep
	tmpDir := t.TempDir()
	moduleFile := filepath.Join(tmpDir, "MODULE.bazel")
	content := `module(
    name = "test",
    version = "1.0.0",
)

bazel_dep(
    name = "rules_go",
    version = "0.42.0",
)
`
	err := os.WriteFile(moduleFile, []byte(content), 0644)
	require.NoError(t, err)

	deps, err := parseModuleDependencies(moduleFile)
	require.NoError(t, err)
	assert.Len(t, deps, 1)
	assert.Equal(t, "rules_go", deps[0].Name)
	assert.Equal(t, "0.42.0", deps[0].Version)
}

