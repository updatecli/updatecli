package stash

import (
	"testing"

	"github.com/updatecli/updatecli/pkg/plugins/resources/stash/client"
)

func TestSummary(t *testing.T) {

	tests := []struct {
		name       string
		repository *Stash
		expected   string
	}{
		{
			name: "Test Summary",
			repository: &Stash{
				Spec: Spec{
					Owner:      "updatecli",
					Repository: "updatecli",
					Branch:     "main",
				},
			},
			expected: "",
		},
		{
			name: "Test Summary with URL",
			repository: &Stash{
				Spec: Spec{
					Owner:      "updatecli",
					Repository: "updatecli",
					Branch:     "main",
					Spec: client.Spec{
						URL: "https://username:password@example.com",
					},
				},
			},
			expected: "example.com/updatecli/updatecli@main",
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
