package utils

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSanitizeFilePathWithWorkingDirectory covers the path containment
// primitive that protects every file based resource against the path
// traversal / arbitrary file write reported in GHSA-hj4x-hm4v-7wpw.
func TestSanitizeFilePathWithWorkingDirectory(t *testing.T) {

	// An OS appropriate absolute path so the absolute-path rejection is
	// exercised identically on POSIX and Windows.
	absolutePath := "/etc/cron.d/evil"
	if runtime.GOOS == "windows" {
		absolutePath = `C:\Windows\Temp\evil`
	}

	testData := []struct {
		name           string
		filePath       string
		workingDir     string
		expectedResult string
		wantErr        bool
	}{
		{
			name:           "relative path stays inside the working directory",
			filePath:       "config/values.yaml",
			workingDir:     "workdir",
			expectedResult: filepath.Join("workdir", "config/values.yaml"),
		},
		{
			name:           "dot slash prefixed relative path is contained",
			filePath:       "./values.yaml",
			workingDir:     "workdir",
			expectedResult: filepath.Join("workdir", "values.yaml"),
		},
		{
			name:           "inner traversal that resolves back inside is allowed",
			filePath:       "sub/../values.yaml",
			workingDir:     "workdir",
			expectedResult: filepath.Join("workdir", "values.yaml"),
		},
		{
			name:       "single dot dot traversal is rejected",
			filePath:   "../evil",
			workingDir: "workdir",
			wantErr:    true,
		},
		{
			name:       "deep dot dot traversal is rejected",
			filePath:   "../../../../etc/cron.d/evil",
			workingDir: "workdir",
			wantErr:    true,
		},
		{
			name:       "traversal hidden behind a subdirectory is rejected",
			filePath:   "sub/../../evil",
			workingDir: "workdir",
			wantErr:    true,
		},
		{
			name:       "absolute path is rejected when a working directory is set",
			filePath:   absolutePath,
			workingDir: "workdir",
			wantErr:    true,
		},
		{
			name:           "absolute path is allowed without a working directory (local run)",
			filePath:       absolutePath,
			workingDir:     "",
			expectedResult: absolutePath,
		},
		{
			name:           "https url is returned untouched",
			filePath:       "https://example.com/values.yaml",
			workingDir:     "workdir",
			expectedResult: "https://example.com/values.yaml",
		},
		{
			name:           "http url is returned untouched",
			filePath:       "http://example.com/values.yaml",
			workingDir:     "workdir",
			expectedResult: "http://example.com/values.yaml",
		},
		{
			name:           "empty working directory keeps the path unchanged",
			filePath:       "values.yaml",
			workingDir:     "",
			expectedResult: "values.yaml",
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := SanitizeFilePathWithWorkingDirectory(tt.filePath, tt.workingDir)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, gotResult)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}

func TestJoinFilePathWithWorkingDirectoryPath(t *testing.T) {

	testData := []struct {
		name           string
		fileName       string
		workingDir     string
		expectedResult string
	}{
		{
			name:           "scenario 1",
			fileName:       "test.yaml",
			workingDir:     "/tmp",
			expectedResult: "/tmp/test.yaml",
		},
		{
			name:           "scenario 2",
			fileName:       "/tmp/test.yaml",
			workingDir:     "/opt",
			expectedResult: "/tmp/test.yaml",
		},
		{
			name:           "scenario 3",
			fileName:       "https://test.yaml",
			workingDir:     "/opt",
			expectedResult: "https://test.yaml",
		},
		{
			name:           "scenario 4",
			fileName:       "http://test.yaml",
			workingDir:     "/opt",
			expectedResult: "http://test.yaml",
		},
		{
			name:           "scenario 5",
			fileName:       "test.yaml",
			workingDir:     "",
			expectedResult: "test.yaml",
		},
		{
			name:           "scenario 6",
			fileName:       "./test.yaml",
			workingDir:     "/opt",
			expectedResult: "/opt/test.yaml",
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := JoinFilePathWithWorkingDirectoryPath(tt.fileName, tt.workingDir)
			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}

// TestFindFilesMatchingPathPattern tests FindFilesMatchingPathPattern function
func TestFindFilesMatchingPathPattern(t *testing.T) {
	testdata := []struct {
		filepath           string
		expectedFoundFiles []string
	}{
		{
			filepath: "*_test.go",
			expectedFoundFiles: []string{
				"fileoperations_test.go",
				"filepath_test.go",
			},
		},
		{
			filepath: "filepath_?est.go",
			expectedFoundFiles: []string{
				"filepath_test.go",
			},
		},
		{
			filepath: "*.go",
			expectedFoundFiles: []string{
				"fileoperations.go",
				"fileoperations_test.go",
				"filepath.go",
				"filepath_test.go",
			},
		},
	}

	for _, data := range testdata {
		t.Run(data.filepath, func(t *testing.T) {
			gotFoundFiles, err := FindFilesMatchingPathPattern("", data.filepath)
			assert.NoError(t, err)
			assert.Equal(t, data.expectedFoundFiles, gotFoundFiles)
		})
	}
}
