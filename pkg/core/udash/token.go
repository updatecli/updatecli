package udash

import (
	"fmt"
	"os"
)

// APIURLSelector selects which stored credential to use when several Udash
// instances are configured. It holds the API URL of the wanted instance and is
// set from the --reportAPI flag.
var APIURLSelector string

// getConfigFromFile return the Udash configuration from the configuration file
func getConfigFromFile(apiURL string) (URL string, ApiURL string, Token string, err error) {
	if APIURLSelector != "" {
		apiURL = APIURLSelector
	}

	data, err := readConfigFile()
	if err != nil {
		return "", "", "", err
	}

	switch apiURL {
	case "":
		authdata, ok := data.Auths[data.Default]
		if ok {
			return authdata.URL, authdata.API, authdata.Token, nil
		}
		return "", "", "", fmt.Errorf("no default token found")
	default:
		authdata, ok := data.Auths[sanitizeTokenID(apiURL)]
		if ok {
			return authdata.URL, authdata.API, authdata.Token, nil
		}
	}

	return "", "", "", fmt.Errorf("token for domain %q not found", apiURL)
}

// getConfigFromEnv return the Udash configuration from environment variables
func getConfigFromEnv() (URL string, ApiURL string, Token string) {
	return os.Getenv(DefaultEnvVariableURL), os.Getenv(DefaultEnvVariableAPIURL), os.Getenv(DefaultEnvVariableAccessToken)
}
