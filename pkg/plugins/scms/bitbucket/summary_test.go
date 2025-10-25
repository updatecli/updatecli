package bitbucket

import (
	"testing"
)

func TestSummary(t *testing.T) {

	tests := []struct {
		name       string
		repository *Bitbucket
		expected   string
	}{
		{
			name: "Test Summary",
			repository: &Bitbucket{
				Spec: Spec{
					Owner:      "updatecli",
					Repository: "updatecli",
					Branch:     "main",
				},
			},
			expected: "bitbucket.org/updatecli/updatecli@main",
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
