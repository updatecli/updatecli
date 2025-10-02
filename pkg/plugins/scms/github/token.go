package github

import (
	"fmt"

	"golang.org/x/oauth2"
)

// getAccessToken retrieves a valid access token from a TokenSource
func getAccessToken(tokenSource oauth2.TokenSource) (string, error) {

	token, err := tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	return token.AccessToken, nil
}
