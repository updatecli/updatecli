package engine

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func setupTestDirectoryStructure(fs afero.Fs, basePath string, structure map[string]bool) error {
	for path, isDir := range structure {
		fullPath := filepath.Join(basePath, path)
		if isDir {
			if err := fs.MkdirAll(fullPath, 0755); err != nil {
				return err
			}
		} else {
			if _, err := fs.Create(fullPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func TestProcessDirectoriesWithPolicyYaml(t *testing.T) {
	fs := afero.NewMemMapFs()

	baseDir := "/test"
	fs.MkdirAll(baseDir, 0755)

	// Define the directory structure for the test
	structure := map[string]bool{
		"dirWithPolicy/Policy.yaml":           false,
		"dirWithoutPolicy/otherfile.txt":      false,
		"emptyDir":                            true,
		"nestedDirWithPolicy/a/b/c":           true,
		"nestedDirWithPolicy/a/b/Policy.yaml": false,
	}

	if err := setupTestDirectoryStructure(fs, baseDir, structure); err != nil {
		t.Fatalf("Failed to setup test directory structure: %s", err)
	}

	tests := []struct {
		name     string
		baseDir  string
		expected bool
	}{
		{"DirWithPolicy", filepath.Join(baseDir, "dirWithPolicy"), true},
		{"DirWithoutPolicy", filepath.Join(baseDir, "dirWithoutPolicy"), false},
		{"EmptyDir", filepath.Join(baseDir, "emptyDir"), false},
		{"NestedDirWithPolicy", filepath.Join(baseDir, "nestedDirWithPolicy"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processDirectoriesWithPolicyYaml(fs, tt.baseDir)
			require.NoError(t, err)
			require.Equal(t, tt.expected, len(got) == 1)
		})
	}
}

func TestVerifyPolicyFiles(t *testing.T) {
	fs := afero.NewMemMapFs()

	dirName := "/policies"
	fileName := "/policies/policy.yaml"
	err := fs.Mkdir(dirName, 0755)
	require.NoError(t, err, "creating directory should not produce an error")

	_, err = fs.Create(fileName)
	require.NoError(t, err, "creating policy file should not produce an error")

	got, err := verifyPolicyFiles(fs, dirName)
	require.Error(t, err, "verifyPolicyFiles should produce an error")
	require.Equal(t, got, false)
}
