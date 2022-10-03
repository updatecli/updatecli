package dasel

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
			fileName:       "test.json",
			workingDir:     "/tmp",
			expectedResult: "/tmp/test.json",
		},
		{
			name:           "scenario 2",
			fileName:       "/tmp/test.json",
			workingDir:     "/opt",
			expectedResult: "/tmp/test.json",
		},
		{
			name:           "scenario 3",
			fileName:       "https://test.json",
			workingDir:     "/opt",
			expectedResult: "https://test.json",
		},
		{
			name:           "scenario 4",
			fileName:       "http://test.json",
			workingDir:     "/opt",
			expectedResult: "http://test.json",
		},
		{
			name:           "scenario 5",
			fileName:       "test.json",
			workingDir:     "",
			expectedResult: "test.json",
		},
		{
			name:           "scenario 6",
			fileName:       "./test.json",
			workingDir:     "/opt",
			expectedResult: "/opt/test.json",
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {

			gotResult := JoinPathWithWorkingDirectoryPath(tt.fileName, tt.workingDir)

			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
