package git

import (
	"testing"
)

func TestSummary(t *testing.T) {

	tests := []struct {
		name       string
		repository *Git
		expected   string
	}{
		{
			name: "Test Summary",
			repository: &Git{
				spec: Spec{
					URL:    "https://example.com/updatecli/updatecli.git",
					Branch: "main",
				},
			},
			expected: "example.com/updatecli/updatecli@main",
		},
		{
			name: "Test Summary with credentials",
			repository: &Git{
				spec: Spec{
					URL:    "https://username:password@example.com/updatecli/updatecli.git",
					Branch: "main",
				},
			},
			expected: "example.com/updatecli/updatecli@main",
		},
		{
			name: "Test summary with git protocol",
			repository: &Git{
				spec: Spec{
					URL:    "git@github.com:updatecli/updatecli.git",
					Branch: "main",
				},
			},
			expected: "git@github.com:updatecli/updatecli.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.repository.Summary()

			if result != tt.expected {
				t.Errorf("Summary() = %v, want %v", result, tt.expected)
			}
		})
	}
}
