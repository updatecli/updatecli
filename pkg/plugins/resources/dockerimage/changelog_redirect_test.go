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
