package json

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
			fileName:       "test.xml",
			workingDir:     "/tmp",
			expectedResult: "/tmp/test.xml",
		},
		{
			name:           "scenario 2",
			fileName:       "/tmp/test.xml",
			workingDir:     "/opt",
			expectedResult: "/tmp/test.xml",
		},
		{
			name:           "scenario 3",
			fileName:       "https://test.xml",
			workingDir:     "/opt",
			expectedResult: "https://test.xml",
		},
		{
			name:           "scenario 4",
			fileName:       "http://test.xml",
			workingDir:     "/opt",
			expectedResult: "http://test.xml",
		},
		{
			name:           "scenario 5",
			fileName:       "test.xml",
			workingDir:     "",
			expectedResult: "test.xml",
		},
		{
			name:           "scenario 6",
			fileName:       "./test.xml",
			workingDir:     "/opt",
			expectedResult: "/opt/test.xml",
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {

			gotResult := joinPathWithWorkingDirectoryPath(tt.fileName, tt.workingDir)

			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
