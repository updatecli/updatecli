package gomodule

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedResult, tt.module.Changelog(tt.from, tt.to))
		})
	}
}
