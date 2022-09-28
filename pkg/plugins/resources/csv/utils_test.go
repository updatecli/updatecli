package csv

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
			fileName:       "test.csv",
			workingDir:     "/tmp",
			expectedResult: "/tmp/test.csv",
		},
		{
			name:           "scenario 2",
			fileName:       "/tmp/test.csv",
			workingDir:     "/opt",
			expectedResult: "/tmp/test.csv",
		},
		{
			name:           "scenario 3",
			fileName:       "https://test.csv",
			workingDir:     "/opt",
			expectedResult: "https://test.csv",
		},
		{
			name:           "scenario 4",
			fileName:       "http://test.csv",
			workingDir:     "/opt",
			expectedResult: "http://test.csv",
		},
		{
			name:           "scenario 5",
			fileName:       "test.csv",
			workingDir:     "",
			expectedResult: "test.csv",
		},
		{
			name:           "scenario 6",
			fileName:       "./test.csv",
			workingDir:     "/opt",
			expectedResult: "/opt/test.csv",
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {

			gotResult := joinPathWithWorkingDirectoryPath(tt.fileName, tt.workingDir)

			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
