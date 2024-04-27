package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
