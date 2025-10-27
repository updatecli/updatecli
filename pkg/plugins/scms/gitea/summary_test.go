package gitea

import (
	"testing"

	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
)

func TestSummary(t *testing.T) {

	tests := []struct {
		name       string
		repository *Gitea
		expected   string
	}{
		{
			name: "Test Summary",
			repository: &Gitea{
				Spec: Spec{
					Owner:      "updatecli",
					Repository: "updatecli",
					Branch:     "main",
				},
			},
			expected: "gitea.com/updatecli/updatecli@main",
		},
		{
			name: "Test Summary with URL",
			repository: &Gitea{
				Spec: Spec{
					Owner:      "updatecli",
					Repository: "updatecli",
					Branch:     "main",
					Spec: client.Spec{
						URL: "https://username:password@gitea.com",
					},
				},
			},
			expected: "gitea.com/updatecli/updatecli@main",
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
