package gitlab

import (
	"testing"

	"github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/client"
)

func TestSummary(t *testing.T) {

	tests := []struct {
		name       string
		repository *Gitlab
		expected   string
	}{
		{
			name: "Test Summary",
			repository: &Gitlab{
				Spec: Spec{
					Owner:      "updatecli",
					Repository: "updatecli",
					Branch:     "main",
				},
			},
			expected: "gitlab.com/updatecli/updatecli@main",
		},
		{
			name: "Test Summary with URL",
			repository: &Gitlab{
				Spec: Spec{
					Owner:      "updatecli",
					Repository: "updatecli",
					Branch:     "main",
					Spec: client.Spec{
						URL: "https://username:password@gitlab.com",
					},
				},
			},
			expected: "gitlab.com/updatecli/updatecli@main",
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
