package github

import "testing"

func TestSummary(t *testing.T) {

	tests := []struct {
		name     string
		github   *Github
		expected string
	}{
		{
			name: "Test Summary",
			github: &Github{
				Spec: Spec{
					Owner:      "updatecli",
					Repository: "updatecli",
					Branch:     "main",
				},
			},
			expected: "github.com/updatecli/updatecli@main",
		},
		{
			name: "Test Summary with URL",
			github: &Github{
				Spec: Spec{
					Owner:      "updatecli",
					Repository: "updatecli",
					Branch:     "main",
					URL:        "https://username:password@github.com",
				},
			},
			expected: "github.com/updatecli/updatecli@main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.github.Summary()

			if result != tt.expected {
				t.Errorf("Summary() = %v, want %v", result, tt.expected)
			}
		})
	}
}
