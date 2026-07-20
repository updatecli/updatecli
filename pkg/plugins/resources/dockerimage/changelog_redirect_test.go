package dockerimage

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedirectToGitHubRawContent(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "rewrites tree to blob",
			input:    "https://github.com/updatecli/policies/tree/main/CHANGELOG.md",
			expected: "https://github.com/updatecli/policies/blob/main/CHANGELOG.md?raw=true",
		},
		{
			name:     "leaves short root-level changelog path untouched",
			input:    "https://github.com/CHANGELOG.md",
			expected: "https://github.com/CHANGELOG.md?raw=true",
		},
		{
			name:     "leaves blob path and adds raw query",
			input:    "https://github.com/updatecli/policies/blob/main/CHANGELOG.md",
			expected: "https://github.com/updatecli/policies/blob/main/CHANGELOG.md?raw=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.input)
			require.NoError(t, err)

			// Must not panic on short paths.
			assert.NotPanics(t, func() {
				redirectToGitHubRawContent(u)
			})

			assert.Equal(t, tt.expected, u.String())
		})
	}
}
