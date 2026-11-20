package dockerimage

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirectToGitHubRawContent(t *testing.T) {
	testdata := []struct {
		name         string
		input        string
		expectedPath string
	}{
		{
			name:         "tree url is rewritten to blob",
			input:        "https://github.com/updatecli/policies/tree/main/CHANGELOG.md",
			expectedPath: "/updatecli/policies/blob/main/CHANGELOG.md",
		},
		{
			name:         "short path without enough segments is left untouched",
			input:        "https://github.com/CHANGELOG.md",
			expectedPath: "/CHANGELOG.md",
		},
		{
			name:         "owner and repo only path is left untouched",
			input:        "https://github.com/updatecli/policies",
			expectedPath: "/updatecli/policies",
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.input)
			require.NoError(t, err)

			redirectToGitHubRawContent(u)

			assert.Equal(t, tt.expectedPath, u.Path)
			assert.Equal(t, "true", u.Query().Get("raw"))
		})
	}
}

func TestRedirectToGiteaRawContent(t *testing.T) {
	testdata := []struct {
		name         string
		input        string
		expectedPath string
	}{
		{
			name:         "branch file page is rewritten to raw",
			input:        "https://codeberg.org/forgejo/forgejo/src/branch/forgejo/RELEASE-NOTES.md",
			expectedPath: "/forgejo/forgejo/raw/branch/forgejo/RELEASE-NOTES.md",
		},
		{
			name:         "tag file page is rewritten to raw",
			input:        "https://gitea.com/gitea/tea/src/tag/v0.9.0/CHANGELOG.md",
			expectedPath: "/gitea/tea/raw/tag/v0.9.0/CHANGELOG.md",
		},
		{
			name:         "raw url is left untouched",
			input:        "https://codeberg.org/owner/repo/raw/branch/main/CHANGELOG.md",
			expectedPath: "/owner/repo/raw/branch/main/CHANGELOG.md",
		},
		{
			name:         "short path is left untouched",
			input:        "https://codeberg.org/CHANGELOG.md",
			expectedPath: "/CHANGELOG.md",
		},
	}

	for _, tt := range testdata {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.input)
			require.NoError(t, err)

			redirectToGiteaRawContent(u)

			assert.Equal(t, tt.expectedPath, u.Path)
		})
	}
}
