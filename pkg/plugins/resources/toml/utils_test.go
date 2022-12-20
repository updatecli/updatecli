package toml

import (
	"testing"

	"gotest.tools/assert"
)

func TestJoinPathWithWorkingDirectoryPath(t *testing.T) {

	testData := []struct {
		name           string
		fileName       string
		workingDir     string
		expectedResult string
	}{
		{
			name:           "scenario 1",
			fileName:       "test.toml",
			workingDir:     "/tmp",
			expectedResult: "/tmp/test.toml",
		},
		{
			name:           "scenario 2",
			fileName:       "/tmp/test.toml",
			workingDir:     "/opt",
			expectedResult: "/tmp/test.toml",
		},
		{
			name:           "scenario 3",
			fileName:       "https://test.toml",
			workingDir:     "/opt",
			expectedResult: "https://test.toml",
		},
		{
			name:           "scenario 4",
			fileName:       "http://test.toml",
			workingDir:     "/opt",
			expectedResult: "http://test.toml",
		},
		{
			name:           "scenario 5",
			fileName:       "test.toml",
			workingDir:     "",
			expectedResult: "test.toml",
		},
		{
			name:           "scenario 6",
			fileName:       "./test.toml",
			workingDir:     "/opt",
			expectedResult: "/opt/test.toml",
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {

			gotResult := joinPathWithWorkingDirectoryPath(tt.fileName, tt.workingDir)

			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
