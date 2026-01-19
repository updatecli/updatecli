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

func TestGetUsername(t *testing.T) {
	tokenValue := "ghp_token"
	emptyString := ""
	oauth2Username := "oauth2"
	customUsername := "custom"

	tests := []struct {
		name             string
		token            *string
		username         *string
		expectedUsername string
	}{
		{
			name:             "No token - returns empty",
			token:            nil,
			username:         nil,
			expectedUsername: "",
		},
		{
			name:             "Empty token - returns empty",
			token:            &emptyString,
			username:         nil,
			expectedUsername: "",
		},
		{
			name:             "Token present, nil username - default to oauth2",
			token:            &tokenValue,
			username:         nil,
			expectedUsername: "oauth2",
		},
		{
			name:             "Token present, empty username",
			token:            &tokenValue,
			username:         &emptyString,
			expectedUsername: "",
		},
		{
			name:             "Token present, oauth2 username",
			token:            &tokenValue,
			username:         &oauth2Username,
			expectedUsername: "oauth2",
		},
		{
			name:             "Token present, custom username",
			token:            &tokenValue,
			username:         &customUsername,
			expectedUsername: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := Terragrunt{
				spec: Spec{
					Token:    tt.token,
					Username: tt.username,
				},
			}

			result := tg.getUsername()
			assert.Equal(t, tt.expectedUsername, result)
		})
	}
}
