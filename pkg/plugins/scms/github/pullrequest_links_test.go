package github

import "testing"

func TestRedirectGitHubPullRequestLinks(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "redirects GitHub pull request URL",
			input:    "See https://github.com/test/repo/pull/123",
			expected: "See https://redirect.github.com/test/repo/pull/123",
		},
		{
			name:     "leaves GitHub release URL unchanged",
			input:    "See https://github.com/test/repo/releases/tag/v1.0.0",
			expected: "See https://github.com/test/repo/releases/tag/v1.0.0",
		},
		{
			name:     "leaves GitHub issue URL unchanged",
			input:    "See https://github.com/test/repo/issues/123",
			expected: "See https://github.com/test/repo/issues/123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := redirectGitHubPullRequestLinks(tt.input)

			if actual != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}
