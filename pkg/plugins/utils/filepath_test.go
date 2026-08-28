package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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
				"resolver_test.go",
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
				"resolver.go",
				"resolver_test.go",
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
