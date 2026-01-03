package terragrunt

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetToken(t *testing.T) {
	emptyString := ""
	githubToken := "ghp_xxxxxxxxxxxx"
	gitlabToken := "glpat-xxxxxxxxxxxx"

	tests := []struct {
		name          string
		token         *string
		expectedToken string
	}{
		{
			name:          "Nil token - no authentication",
			token:         nil,
			expectedToken: "",
		},
		{
			name:          "Empty token - no authentication",
			token:         &emptyString,
			expectedToken: "",
		},
		{
			name:          "GitHub token - use specific token",
			token:         &githubToken,
			expectedToken: "ghp_xxxxxxxxxxxx",
		},
		{
			name:          "GitLab token - use specific token",
			token:         &gitlabToken,
			expectedToken: "glpat-xxxxxxxxxxxx",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test instance
			tg := Terragrunt{
				spec: Spec{
					Token: tt.token,
				},
			}

			// Test
			result := tg.getToken()
			assert.Equal(t, tt.expectedToken, result)
		})
	}
}
