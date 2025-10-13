package token

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/app"
	"golang.org/x/oauth2"
)

func TestGetAccessToken(t *testing.T) {
	expectedToken := "testtoken"
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: expectedToken})

	token, err := GetAccessToken(tokenSource)
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, token)

	// Simulate error from TokenSource
	badTokenSource := oauth2.StaticTokenSource(nil)
	_, err = GetAccessToken(badTokenSource)
	assert.Error(t, err)
}

func TestGetTokenSourceFromEnv_PAT(t *testing.T) {
	os.Setenv("UPDATECLI_GITHUB_TOKEN", "pat_token")
	defer os.Unsetenv("UPDATECLI_GITHUB_TOKEN")

	os.Setenv("UPDATECLI_GITHUB_USERNAME", "")
	defer os.Unsetenv("UPDATECLI_GITHUB_USERNAME")

	username, tokenSource, err := GetTokenSourceFromEnv()
	assert.NoError(t, err)
	assert.Equal(t, githubTokenUsername, username)
	token, err := tokenSource.Token()
	assert.NoError(t, err)
	assert.Equal(t, "pat_token", token.AccessToken)
}

// This test checks the behavior when GitHub App environment variables are set but invalid
func TestGetTokenSourceFromEnv_GitHubApp(t *testing.T) {
	os.Unsetenv("UPDATECLI_GITHUB_TOKEN")
	os.Setenv("UPDATECLI_GITHUB_APP_CLIENT_ID", "123456")
	os.Setenv("UPDATECLI_GITHUB_APP_PRIVATE_KEY", "dummykey")
	os.Setenv("UPDATECLI_GITHUB_APP_INSTALLATION_ID", "789012")
	os.Setenv("UPDATECLI_GITHUB_APP_EXPIRATION_TIME", "3600")
	defer func() {
		os.Unsetenv("UPDATECLI_GITHUB_APP_CLIENT_ID")
		os.Unsetenv("UPDATECLI_GITHUB_APP_PRIVATE_KEY")
		os.Unsetenv("UPDATECLI_GITHUB_APP_INSTALLATION_ID")
		os.Unsetenv("UPDATECLI_GITHUB_APP_EXPIRATION_TIME")
	}()

	username, tokenSource, err := GetTokenSourceFromEnv()
	assert.ErrorContains(t, err, "failed to get oauth2 token source from GitHub App spec: creating GitHub App token source: invalid key: Key must be a PEM encoded PKCS1 or PKCS8 key")
	assert.Equal(t, "", username)
	assert.Nil(t, tokenSource)
}

func TestGetTokenSourceFromEnv_None(t *testing.T) {
	os.Unsetenv("UPDATECLI_GITHUB_TOKEN")
	os.Unsetenv("UPDATECLI_GITHUB_APP_CLIENT_ID")
	os.Unsetenv("UPDATECLI_GITHUB_APP_PRIVATE_KEY")
	os.Unsetenv("UPDATECLI_GITHUB_APP_INSTALLATION_ID")
	os.Unsetenv("UPDATECLI_GITHUB_APP_EXPIRATION_TIME")

	username, tokenSource, err := GetTokenSourceFromEnv()
	assert.NoError(t, err)
	assert.Empty(t, username)
	assert.Nil(t, tokenSource)
}

func TestGetFallbackTokenSourceFromEnv(t *testing.T) {
	os.Setenv("GITHUB_TOKEN", "fallback_token")
	defer os.Unsetenv("GITHUB_TOKEN")

	username, tokenSource := GetFallbackTokenSourceFromEnv()
	assert.Equal(t, githubTokenUsername, username)
	token, err := tokenSource.Token()
	assert.NoError(t, err)
	assert.Equal(t, "fallback_token", token.AccessToken)

	os.Unsetenv("GITHUB_TOKEN")
	username, tokenSource = GetFallbackTokenSourceFromEnv()
	assert.Empty(t, username)
	assert.Nil(t, tokenSource)
}

func TestGetTokenSourceFromConfig_PAT(t *testing.T) {
	username, tokenSource, err := GetTokenSourceFromConfig("user", "pat_token", nil)
	assert.NoError(t, err)
	assert.Equal(t, "user", username)
	token, err := tokenSource.Token()
	assert.NoError(t, err)
	assert.Equal(t, "pat_token", token.AccessToken)
}

func TestGetTokenSourceFromConfig_GitHubApp(t *testing.T) {
	appSpec := &app.Spec{
		ClientID:       "123456",
		PrivateKey:     "dummykey",
		InstallationID: "789012",
		ExpirationTime: "3600",
	}
	username, tokenSource, err := GetTokenSourceFromConfig("user", "", appSpec)
	assert.ErrorContains(t, err, "failed to get oauth2 token source from GitHub App spec: creating GitHub App token source: invalid key: Key must be a PEM encoded PKCS1 or PKCS8 key")
	assert.Equal(t, "", username)
	assert.Nil(t, tokenSource)
}

func TestGetTokenSourceFromConfig_None(t *testing.T) {
	username, tokenSource, err := GetTokenSourceFromConfig("user", "", nil)
	assert.NoError(t, err)
	assert.Empty(t, username)
	assert.Nil(t, tokenSource)
}
