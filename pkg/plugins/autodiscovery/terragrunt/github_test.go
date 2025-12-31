package terragrunt

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGitHubToken(t *testing.T) {
	tests := []struct {
		name          string
		specToken     string
		envToken      string
		ghToken       string
		expectedToken string
	}{
		{
			name:          "Token from spec takes precedence",
			specToken:     "spec-token",
			envToken:      "env-token",
			ghToken:       "gh-token",
			expectedToken: "spec-token",
		},
		{
			name:          "Token from UPDATECLI_GITHUB_TOKEN when spec is empty",
			specToken:     "",
			envToken:      "updatecli-token",
			ghToken:       "gh-token",
			expectedToken: "updatecli-token",
		},
		{
			name:          "Token from GITHUB_TOKEN when others are empty",
			specToken:     "",
			envToken:      "",
			ghToken:       "github-token",
			expectedToken: "github-token",
		},
		{
			name:          "Empty token when all are empty",
			specToken:     "",
			envToken:      "",
			ghToken:       "",
			expectedToken: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original env
			origUpdatecliToken := os.Getenv("UPDATECLI_GITHUB_TOKEN")
			origGHToken := os.Getenv("GITHUB_TOKEN")

			// Clean env
			os.Unsetenv("UPDATECLI_GITHUB_TOKEN")
			os.Unsetenv("GITHUB_TOKEN")

			// Set test env
			if tt.envToken != "" {
				os.Setenv("UPDATECLI_GITHUB_TOKEN", tt.envToken)
			}
			if tt.ghToken != "" {
				os.Setenv("GITHUB_TOKEN", tt.ghToken)
			}

			// Create test instance
			tg := Terragrunt{
				spec: Spec{
					GitHub: GitHubSpec{
						Token: tt.specToken,
					},
				},
			}

			// Test
			result := tg.getGitHubToken()
			assert.Equal(t, tt.expectedToken, result)

			// Restore original env
			os.Unsetenv("UPDATECLI_GITHUB_TOKEN")
			os.Unsetenv("GITHUB_TOKEN")
			if origUpdatecliToken != "" {
				os.Setenv("UPDATECLI_GITHUB_TOKEN", origUpdatecliToken)
			}
			if origGHToken != "" {
				os.Setenv("GITHUB_TOKEN", origGHToken)
			}
		})
	}
}
