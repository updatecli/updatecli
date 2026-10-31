package credential

import (
	"os"

	"github.com/updatecli/updatecli/pkg/core/log"
)

var (
	// ENVIRONMENT_GITEA_USERNAME is the environment variable used to set the Gitea username.
	ENVIRONMENT_GITEA_USERNAME string = "UPDATECLI_GITEA_USERNAME"
	// ENVIRONMENT_GITEA_TOKEN is the environment variable used to set the Gitea token.
	ENVIRONMENT_GITEA_TOKEN string = "UPDATECLI_GITEA_TOKEN" // #nosec G101 -- This is not a hardcoded credential.
)

// GetCredentialsFromEnv retrieves the Gitea username and token from environment variables.
func GetCredentialsFromEnv() (username string, token string) {
	username = os.Getenv(ENVIRONMENT_GITEA_USERNAME)
	if username != "" {
		log.Debugf("Gitea username found in environment variable %s", ENVIRONMENT_GITEA_USERNAME)
	}
	token = os.Getenv(ENVIRONMENT_GITEA_TOKEN)
	if token != "" {
		log.Debugf("Gitea token found in environment variable %s", ENVIRONMENT_GITEA_TOKEN)
	}
	return username, token
}
