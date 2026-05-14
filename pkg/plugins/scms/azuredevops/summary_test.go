package azuredevops

import (
	"testing"

	azdoclient "github.com/updatecli/updatecli/pkg/plugins/resources/azuredevops/client"
)

func TestSummary(t *testing.T) {
	tests := []struct {
		name       string
		repository *AzureDevOps
		expected   string
	}{
		{
			name: "Test Summary",
			repository: &AzureDevOps{
				Spec: Spec{
					Branch: "main",
					Spec: azdoclient.Spec{
						URL:        "https://dev.azure.com/updatecli",
						Project:    "updatecli",
						Repository: "updatecli",
					},
				},
			},
			expected: "dev.azure.com/updatecli/updatecli/_git/updatecli@main",
		},
		{
			name: "Test Summary with URL credentials",
			repository: &AzureDevOps{
				Spec: Spec{
					Branch: "main",
					Spec: azdoclient.Spec{
						// #nosec G101
						URL:        "https://username:password@dev.azure.com/updatecli",
						Project:    "updatecli",
						Repository: "updatecli",
					},
				},
			},
			expected: "dev.azure.com/updatecli/updatecli/_git/updatecli@main",
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
