package pypi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithVPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "Empty string unchanged", input: "", expected: ""},
		{name: "Already prefixed unchanged", input: "v2.31.0", expected: "v2.31.0"},
		{name: "Plain version gets v prefix", input: "2.31.0", expected: "v2.31.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, withVPrefix(tt.input))
		})
	}
}

func TestChangelogFromGitHub(t *testing.T) {
	tests := []struct {
		name      string
		rawURL    string
		from      string
		to        string
		expectNil bool
	}{
		{
			name:      "Non-GitHub URL returns nil",
			rawURL:    "https://gitlab.com/owner/repo",
			from:      "1.0.0",
			to:        "1.0.0",
			expectNil: true,
		},
		{
			name:      "Malformed URL with too few path parts returns nil",
			rawURL:    "https://github.com/onlyowner",
			from:      "1.0.0",
			to:        "1.0.0",
			expectNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := changelogFromGitHub(tt.rawURL, tt.from, tt.to)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

func TestChangelogFromPyPI(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		baseURL     string
		from        string
		to          string
		expectNil   bool
		expectURLs  []string
	}{
		{
			name:        "Same from and to produces single entry",
			packageName: "requests",
			baseURL:     "https://pypi.org/",
			from:        "2.31.0",
			to:          "2.31.0",
			expectURLs:  []string{"https://pypi.org/pypi/requests/2.31.0/"},
		},
		{
			name:        "Different from and to produces two entries",
			packageName: "requests",
			baseURL:     "https://pypi.org/",
			from:        "2.28.0",
			to:          "2.31.0",
			expectURLs: []string{
				"https://pypi.org/pypi/requests/2.28.0/",
				"https://pypi.org/pypi/requests/2.31.0/",
			},
		},
		{
			name:        "Empty from and to returns nil",
			packageName: "requests",
			baseURL:     "https://pypi.org/",
			from:        "",
			to:          "",
			expectNil:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := changelogFromPyPI(tt.packageName, tt.baseURL, tt.from, tt.to)
			if tt.expectNil {
				assert.Nil(t, result)
				return
			}
			require.NotNil(t, result)
			assert.Equal(t, len(tt.expectURLs), len(*result))
			for i, expectedURL := range tt.expectURLs {
				assert.Equal(t, expectedURL, (*result)[i].URL)
			}
		})
	}
}

func TestChangelog_FallsBackToPyPI(t *testing.T) {
	// Package with no Source URL in project_urls falls back to PyPI release pages.
	const noGitHubData = `{
  "info": {
    "name": "requests",
    "version": "2.31.0",
    "project_urls": {}
  },
  "releases": {
    "2.31.0": [{"yanked": false}]
  }
}`

	spec := Spec{
		Name:  "requests",
		URL:   "https://pypi.example.com",
		Token: "validtoken",
	}
	p, err := New(spec)
	require.NoError(t, err)
	p.webClient = GetMockClient("https://pypi.example.com/", "validtoken", noGitHubData, 200)

	changelogs := p.Changelog("2.31.0", "2.31.0")
	require.NotNil(t, changelogs)
	assert.Equal(t, 1, len(*changelogs))
	assert.Contains(t, (*changelogs)[0].URL, "requests")
	assert.Contains(t, (*changelogs)[0].URL, "2.31.0")
}
