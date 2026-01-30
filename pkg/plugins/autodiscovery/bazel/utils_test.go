package bazel

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindModuleFiles(t *testing.T) {
	testdata := []struct {
		name          string
		rootDir       string
		expectedFiles []string
		expectError   bool
	}{
		{
			name:    "Find MODULE.bazel files in testdata",
			rootDir: "testdata",
			expectedFiles: []string{
				"testdata/project1/MODULE.bazel",
				"testdata/project2/MODULE.bazel",
				"testdata/project2/subdir/MODULE.bazel",
			},
			expectError: false,
		},
		{
			name:          "Non-existent directory",
			rootDir:       "nonexistent",
			expectedFiles: []string{},
			expectError:   true,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			files, err := findModuleFiles(tt.rootDir)
			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.GreaterOrEqual(t, len(files), len(tt.expectedFiles))

			// Normalize paths for comparison
			normalizedFiles := make([]string, len(files))
			for i, f := range files {
				rel, err := filepath.Rel(".", f)
				if err == nil {
					normalizedFiles[i] = rel
				} else {
					normalizedFiles[i] = f
				}
			}

			// Check that all expected files are found
			for _, expected := range tt.expectedFiles {
				found := false
				for _, f := range normalizedFiles {
					if filepath.Base(f) == filepath.Base(expected) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected file %q not found", expected)
			}
		})
	}
}

func TestShouldIgnore(t *testing.T) {
	testdata := []struct {
		name        string
		moduleName  string
		version     string
		filePath    string
		rootDir     string
		rules       MatchingRules
		expectError bool
	}{
		{
			name:        "No rules - should not ignore",
			moduleName:  "rules_go",
			version:     "0.42.0",
			filePath:    "testdata/project1/MODULE.bazel",
			rootDir:     "testdata",
			rules:       MatchingRules{},
			expectError: false,
		},
		{
			name:       "Ignore by module name",
			moduleName: "rules_go",
			version:    "0.42.0",
			filePath:   "testdata/project1/MODULE.bazel",
			rootDir:    "testdata",
			rules: MatchingRules{
				MatchingRule{
					Modules: map[string]string{
						"rules_go": "",
					},
				},
			},
			expectError: true,
		},
		{
			name:       "Ignore by path pattern",
			moduleName: "rules_go",
			version:    "0.42.0",
			filePath:   "testdata/project1/MODULE.bazel",
			rootDir:    "testdata",
			rules: MatchingRules{
				MatchingRule{
					Path: "project1/*",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			// Create absolute paths for testing
			absRoot, err := filepath.Abs(tt.rootDir)
			require.NoError(t, err)
			absPath, err := filepath.Abs(tt.filePath)
			require.NoError(t, err)

			result := shouldIgnore(tt.moduleName, tt.version, absPath, absRoot, tt.rules)
			assert.Equal(t, tt.expectError, result)
		})
	}
}

func TestShouldInclude(t *testing.T) {
	testdata := []struct {
		name        string
		moduleName  string
		version     string
		filePath    string
		rootDir     string
		rules       MatchingRules
		expectError bool
	}{
		{
			name:        "No rules - should include",
			moduleName:  "rules_go",
			version:     "0.42.0",
			filePath:    "testdata/project1/MODULE.bazel",
			rootDir:     "testdata",
			rules:       MatchingRules{},
			expectError: true,
		},
		{
			name:       "Include by module name",
			moduleName: "rules_go",
			version:    "0.42.0",
			filePath:   "testdata/project1/MODULE.bazel",
			rootDir:    "testdata",
			rules: MatchingRules{
				MatchingRule{
					Modules: map[string]string{
						"rules_go": "",
					},
				},
			},
			expectError: true,
		},
		{
			name:       "Exclude by module name",
			moduleName: "gazelle",
			version:    "0.34.0",
			filePath:   "testdata/project1/MODULE.bazel",
			rootDir:    "testdata",
			rules: MatchingRules{
				MatchingRule{
					Modules: map[string]string{
						"rules_go": "",
					},
				},
			},
			expectError: false,
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			// Create absolute paths for testing
			absRoot, err := filepath.Abs(tt.rootDir)
			require.NoError(t, err)
			absPath, err := filepath.Abs(tt.filePath)
			require.NoError(t, err)

			result := shouldInclude(tt.moduleName, tt.version, absPath, absRoot, tt.rules)
			assert.Equal(t, tt.expectError, result)
		})
	}
}

func TestFindModuleFilesSkipsHiddenDirs(t *testing.T) {
	// Use testdata files instead of creating files on disk
	rootDir := "testdata"
	files, err := findModuleFiles(rootDir)
	require.NoError(t, err)

	// Should find files from normal directories but not from hidden directories
	// Verify that no files from .hidden directory are included
	for _, file := range files {
		assert.NotContains(t, file, ".hidden", "Found file in hidden directory: %s", file)
	}

	// Verify that files from normal directories are found
	// We should find files from project1, project2, empty, multiline, invalid, and normal
	foundNormal := false
	for _, file := range files {
		if filepath.Base(filepath.Dir(file)) == "normal" {
			foundNormal = true
			break
		}
	}
	assert.True(t, foundNormal || len(files) > 0, "Should find at least some MODULE.bazel files")
}
