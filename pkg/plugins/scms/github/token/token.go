package token

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/app"
	"golang.org/x/oauth2"
)

var (
	githubTokenUsername  string = "oauth2"
	githubAccessTokenKey string = "x-access-token"
)

// GetAccessToken retrieves a valid access token from a TokenSource
func GetAccessToken(tokenSource oauth2.TokenSource) (string, error) {

	token, err := tokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	if token == nil || token.AccessToken == "" {
		return "", fmt.Errorf("no access token found")
	}

	return token.AccessToken, nil
}

// GetTokenSourceFromEnv retrieves a valid access token from environment variables
// It supports both personal access tokens and GitHub App tokens
// The precedence is as follows:
//  1. "UPDATECLI_GITHUB_TOKEN"
//  2. GitHub App environment variables (GITHUB_APP_ID, GITHUB_APP_INSTALLATION_ID, GITHUB_APP_PRIVATE_KEY)
func GetTokenSourceFromEnv() (string, oauth2.TokenSource, error) {

	if token := os.Getenv("UPDATECLI_GITHUB_TOKEN"); token != "" {

		logrus.Debugf("using GitHub token from environment variable UPDATECLI_GITHUB_TOKEN")

		username := os.Getenv("UPDATECLI_GITHUB_USERNAME")
		if username == "" {
			username = githubTokenUsername
		}

		return username, oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: token,
			}), nil
	}

	GitHubAppSpecFromEnv := app.NewSpecFromEnv()
	if GitHubAppSpecFromEnv != nil {
		logrus.Debugf("using GitHub App authentication from environment variables")
		tokenSource, err := GitHubAppSpecFromEnv.Getoauth2TokenSource()
		if err != nil {
			return "", nil, fmt.Errorf("failed to get oauth2 token source from GitHub App spec: %w", err)
		}
		return githubAccessTokenKey, tokenSource, nil

	}

	return "", nil, nil
}

// GetFallbackTokenSourceFromEnv retrieves a valid access token from the "GITHUB_TOKEN" environment variable
// This function is used as a fallback mechanism if no other token source could be found
// It is primarily intended to support GitHub Actions workflows
func GetFallbackTokenSourceFromEnv() (string, oauth2.TokenSource) {
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		logrus.Debugf("using GitHub token from environment variable GITHUB_TOKEN")

		return githubTokenUsername, oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: token,
			})
	}
	return "", nil
}

// GetTokenSourceFromConfig retrieves a valid access token from a Spec configuration
// It supports both personal access tokens and GitHub App tokens
// The precedence is as follows:
//  1. Token provided in the Spec configuration
//
// 2. GitHub App configuration in the Spec
// 3. No token found, return an error
func GetTokenSourceFromConfig(username, token string, app *app.Spec) (string, oauth2.TokenSource, error) {

	if token != "" {
		logrus.Debugf("using GitHub token from configuration")

		return username, oauth2.StaticTokenSource(
			&oauth2.Token{
				AccessToken: token,
			}), nil
	}

	if app != nil {
		logrus.Debugf("using GitHub App authentication from configuration")
		tokenSource, err := app.Getoauth2TokenSource()
		if err != nil {
			return "", nil, fmt.Errorf("failed to get oauth2 token source from GitHub App spec: %w", err)
		}
		return githubAccessTokenKey, tokenSource, nil
	}

	return "", nil, nil
}
