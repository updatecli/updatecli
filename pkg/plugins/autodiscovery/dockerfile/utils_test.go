package dockerfile

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSearchFiles(t *testing.T) {
	testdata := []struct {
		name          string
		rootDir       string
		expectedfiles []string
	}{
		{
			name:    "Nominal case",
			rootDir: "test/testdata/",
			expectedfiles: []string{
				"test/testdata/Dockerfile",
				"test/testdata/alpine/Dockerfile",
				"test/testdata/jenkins/Dockerfile",
				"test/testdata/updatecli-action/Dockerfile",
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			gotFiles, err := searchDockerfiles(
				"test/testdata/", DefaultFileMatch[:])
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedfiles, gotFiles)
		})
	}
}

func TestGetDockerfileData(t *testing.T) {
	testdata := []struct {
		name                string
		filepath            string
		expectedInstruction []instruction
	}{
		{
			name:     "Default case",
			filepath: "test/testdata/Dockerfile",
			expectedInstruction: []instruction{
				{
					name:  "FROM",
					value: "updatecli/updatecli:v0.37.0",
					image: "updatecli/updatecli:v0.37.0",
				},
				{
					name:  "FROM",
					value: "updatecli/updatecli:v0.38.0",
					image: "updatecli/updatecli:v0.38.0",
				},
				{
					name:  "FROM",
					value: "updatecli/updatecli:v0.36.0",
					image: "updatecli/updatecli:v0.36.0",
				},
				{
					name:          "ARG",
					value:         "alpine_version",
					image:         "alpine:3.16.3",
					trimArgPrefix: "alpine:",
				},
			},
		},
		{
			name:     "Alpine case with ARG",
			filepath: "test/testdata/alpine/Dockerfile",
			expectedInstruction: []instruction{
				{
					name:          "ARG",
					value:         "alpine_version",
					image:         "alpine:3.16.3",
					trimArgPrefix: "alpine:",
					arch:          "ppc64",
				},
				{
					name:          "ARG",
					value:         "debian_version",
					image:         "debian:8",
					trimArgPrefix: "debian:",
					arch:          "arch64",
				},
				{
					name:  "FROM",
					value: "opensuse:15.4",
					image: "opensuse:15.4",
					arch:  "ppc64",
				},
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			gotInstructions, err := parseDockerfile(tt.filepath)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedInstruction, gotInstructions)
		})
	}
}

func TestExtractArgName(t *testing.T) {
	testdata := []struct {
		name                 string
		input                string
		expectedPrefix       string
		expectedArgName      string
		expectedSuffix       string
		expectedError        bool
		expectedErrorMessage error
	}{
		{
			name:            "Default case",
			input:           "${alpine_version}",
			expectedPrefix:  "",
			expectedArgName: "alpine_version",
			expectedSuffix:  "",
		},
		{
			name:            "Default case",
			input:           "2.235-lts",
			expectedPrefix:  "",
			expectedArgName: "",
			expectedSuffix:  "",
		},
		{
			name:            "Default case",
			input:           "${jenkins_version}-lts",
			expectedPrefix:  "",
			expectedArgName: "jenkins_version",
			expectedSuffix:  "-lts",
		},
		{
			name:            "Default case",
			input:           "lts-${jenkins_version}",
			expectedPrefix:  "lts-",
			expectedArgName: "jenkins_version",
			expectedSuffix:  "",
		},
		{
			name:                 "Default case",
			input:                "${jenkins_release_type}-${jenkins_version}",
			expectedPrefix:       "lts-",
			expectedArgName:      "jenkins_version",
			expectedSuffix:       "",
			expectedError:        true,
			expectedErrorMessage: errors.New("more than one variable detected in the Dockerfile instruction. Updatecli do not support this at the moment"),
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			gotPrefix, gotArgName, gotSuffix, gotErr := extractArgName(tt.input)

			if tt.expectedError {
				assert.EqualError(t, tt.expectedErrorMessage, gotErr.Error())
			} else {
				assert.NoError(t, gotErr)
				assert.Equal(t, tt.expectedPrefix, gotPrefix)
				assert.Equal(t, tt.expectedArgName, gotArgName)
				assert.Equal(t, tt.expectedSuffix, gotSuffix)
			}
		})
	}
}
