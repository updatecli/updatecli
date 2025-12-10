package dockerfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/plugins/resources/dockerfile/simpletextparser/keywords"
)

func TestSearchFiles(t *testing.T) {
	testdata := []struct {
		name          string
		rootDir       string
		expectedfiles []string
	}{
		{
			name:    "Nominal case",
			rootDir: "testdata/",
			expectedfiles: []string{
				"testdata/Dockerfile",
				"testdata/alpine/Dockerfile",
				"testdata/jenkins/Dockerfile",
				"testdata/multi-variable/Dockerfile",
				"testdata/python-slim/Dockerfile",
				"testdata/scratch-and-base/Dockerfile",
				"testdata/similar-stage-and-image/Dockerfile",
				"testdata/updatecli-action/Dockerfile",
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			gotFiles, err := searchDockerfiles(
				"testdata/", DefaultFileMatch[:])
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedfiles, gotFiles)
		})
	}
}

func TestGetDockerfileData(t *testing.T) {
	testdata := []struct {
		name                string
		filepath            string
		expectedInstruction []keywords.FromToken
		expectedArgs        map[string]keywords.SimpleTokens
	}{
		{
			name:     "Default case",
			filepath: "testdata/Dockerfile",
			expectedInstruction: []keywords.FromToken{
				{
					Keyword:  "FROM",
					Image:    "updatecli/updatecli",
					Tag:      "v0.37.0",
					Platform: "${BUILDPLATFORM}",
					Args: map[string]*keywords.FromTokenArgs{
						"platform": {
							Name: "BUILDPLATFORM",
						},
					},
				},
				{
					Keyword: "FROM",
					Image:   "updatecli/updatecli",
					Tag:     "v0.38.0",
				},
				{
					Keyword: "FROM",
					Image:   "updatecli/updatecli",
					Tag:     "v0.36.0",
					Alias:   "builder",
					AliasKw: "as",
				},
				{
					Keyword: "FROM",
					Image:   "alpine",
					Tag:     "${alpine_version}",
					Alias:   "base",
					AliasKw: "AS",
					Args: map[string]*keywords.FromTokenArgs{
						"tag": {
							Name: "alpine_version",
						},
					},
				},
			},
			expectedArgs: map[string]keywords.SimpleTokens{
				"alpine_version": {
					Keyword: "ARG",
					Name:    "alpine_version",
					Value:   "3.16.3",
				},
			},
		},
		{
			name:     "Alpine case with ARG",
			filepath: "testdata/alpine/Dockerfile",
			expectedInstruction: []keywords.FromToken{
				{
					Keyword:  "FROM",
					Platform: "linux/ppc64",
					Image:    "alpine",
					Tag:      "${alpine_version}",
					Alias:    "base_alpine",
					AliasKw:  "AS",
					Args: map[string]*keywords.FromTokenArgs{
						"tag": {
							Name: "alpine_version",
						},
					},
				},
				{
					Keyword:  "FROM",
					Platform: "${platform}",
					Image:    "debian",
					Tag:      "${debian_version}",
					Args: map[string]*keywords.FromTokenArgs{
						"platform": {
							Name: "platform",
						},
						"tag": {
							Name: "debian_version",
						},
					},
				},
				{
					Keyword:  "FROM",
					Platform: "windows/ppc64",
					Image:    "opensuse",
					Tag:      "15.4",
				},
				{
					Keyword:  "FROM",
					Platform: "linux/ppc64",
					Image:    "alpine",
					Tag:      "${alpine_version}${debian_version}",
					Alias:    "alpine",
					AliasKw:  "AS",
				},
			},
			expectedArgs: map[string]keywords.SimpleTokens{
				"alpine_version": {
					Keyword: "ARG",
					Name:    "alpine_version",
					Value:   "3.16.3",
				},
				"debian_version": {
					Keyword: "ARG",
					Name:    "debian_version",
					Value:   "8",
				},
				"platform": {
					Keyword: "ARG",
					Name:    "platform",
					Value:   "linux/arch64",
				},
			},
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			gotInstructions, gotArgs, err := parseDockerfile(tt.filepath)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedInstruction, gotInstructions)
			assert.Equal(t, tt.expectedArgs, gotArgs)
		})
	}
}
