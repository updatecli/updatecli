package gomodule

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestChangelog(t *testing.T) {
	tests := []struct {
		name           string
		from           string
		to             string
		module         GoModule
		expectedResult *result.Changelogs
	}{
		{
			name: "Test getting changelog from github",
			from: "v0.0.1",
			to:   "v0.0.1",
			module: GoModule{
				Spec: Spec{
					Module: "github.com/updatecli/updatecli",
				},
			},
			expectedResult: &result.Changelogs{
				{
					Title:       "v0.0.1",
					URL:         "https://github.com/updatecli/updatecli/releases/tag/v0.0.1",
					Body:        "## Changes\r\n\r\n- Add github repository type to target stage @olblak (#4)\r\n\r\n## 🚀 Features\r\n\r\n- Add Docker Image @olblak (#3)\r\n\r\n## 🐛 Bug Fixes\r\n\r\n- Rename release-drafter config file @olblak (#5)\r\n\r\n## Contributors\r\n\r\n@olblak\r\n",
					PublishedAt: "2020-02-19 20:34:42 +0000 UTC",
				},
			},
		},
		{
			name: "Test getting changelog from helm.sh/helm/v3",
			from: "v3.17.1",
			to:   "v3.17.1",
			module: GoModule{
				Spec: Spec{
					Module: "helm.sh/helm/v3",
				},
			},
			expectedResult: &result.Changelogs{
				{
					Title:       "v3.17.1",
					URL:         "https://github.com/helm/helm/releases/tag/v3.17.1",
					PublishedAt: "2025-02-12 21:01:05 +0000 UTC",
				},
			},
		},
		{
			name: "Test do not exist module",
			from: "v1.67.0",
			to:   "v1.67.0",
			module: GoModule{
				Spec: Spec{
					Module: "donotexit.com/ini.v1",
				},
			},
			expectedResult: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResultPtr := tt.module.Changelog(tt.from, tt.to)

			if tt.expectedResult == nil && gotResultPtr != nil {
				t.Fail()
			} else if tt.expectedResult == nil && gotResultPtr == nil {
				return
			} else {
				require.NotNil(t, gotResultPtr)
			}

			gotResult := *gotResultPtr

			require.Equal(t, len(*tt.expectedResult), len(gotResult))
			for i := range *tt.expectedResult {
				expectedResult := *tt.expectedResult
				assert.Equal(t, expectedResult[i].Title, gotResult[i].Title)
				assert.Equal(t, expectedResult[i].URL, gotResult[i].URL)
				assert.Equal(t, expectedResult[i].PublishedAt, gotResult[i].PublishedAt)
			}
		})
	}
}

func TestGetSourceURL(t *testing.T) {

	testdata := []struct {
		name        string
		htmlContent string
		expectedURL string
	}{
		{
			name: "working scenario",
			htmlContent: `<html><head>
      <meta name="go-import"
            content="sigs.k8s.io/yaml
                     git https://github.com/kubernetes-sigs/yaml">
      <meta name="go-source"
            content="sigs.k8s.io/yaml
                     https://github.com/kubernetes-sigs/yaml
                     https://github.com/kubernetes-sigs/yaml/tree/master{/dir}
                     https://github.com/kubernetes-sigs/yaml/blob/master{/dir}/{file}#L{line}">
</head></html>
`,
			expectedURL: "github.com/kubernetes-sigs/yaml",
		},
		{
			name: "empty html content",
			htmlContent: `<html><head></head></html>
`,
			expectedURL: "",
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := getGitRepositoryURL(tt.htmlContent)
			assert.Equal(t, tt.expectedURL, gotResult)
		})
	}
}
