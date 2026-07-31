package credential

import (
	"os"

	"github.com/sirupsen/logrus"
)

var (
	// ENVIRONMENT_AZURE_DEVOPS_USERNAME is the environment variable used to set the Azure DevOps username.
	ENVIRONMENT_AZURE_DEVOPS_USERNAME string = "UPDATECLI_AZURE_DEVOPS_USERNAME"
	// ENVIRONMENT_AZURE_DEVOPS_TOKEN is the environment variable used to set the Azure DevOps token.
	ENVIRONMENT_AZURE_DEVOPS_TOKEN string = "UPDATECLI_AZURE_DEVOPS_TOKEN" // #nosec G101 -- This is not a hardcoded credential.
)

// GetCredentialsFromEnv retrieves the Azure DevOps username and token from environment variables.
func GetCredentialsFromEnv() (username string, token string) {
	username = os.Getenv(ENVIRONMENT_AZURE_DEVOPS_USERNAME)
	if username != "" {
		logrus.Debugf("Azure DevOps username found in environment variable %s", ENVIRONMENT_AZURE_DEVOPS_USERNAME)
	}
	token = os.Getenv(ENVIRONMENT_AZURE_DEVOPS_TOKEN)
	if token != "" {
		logrus.Debugf("Azure DevOps token found in environment variable %s", ENVIRONMENT_AZURE_DEVOPS_TOKEN)
	}
	return username, token
}
