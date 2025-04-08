package udash

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Login will open a browser to authenticate a user and retrieve an access token
func Login(udashEndpoint, udashAPIEndpoint, clientID, issuer, audience, accessToken string) error {

	if udashAPIEndpoint == "" {
		udashAPIEndpoint = strings.TrimSuffix(udashEndpoint, "/") + "/api"
	}

	envVariableURL := os.Getenv(DefaultEnvVariableURL)
	envVariableAccessToken := os.Getenv(DefaultEnvVariableAccessToken)
	envVariableAPIURL := os.Getenv(DefaultEnvVariableAPIURL)

	setParam := func(flagParam *string, envParam, flagParamName, envParamName string) {
		if *flagParam != "" && envParam != "" {
			logrus.Debugf("%s provided via flag and environment variable %q, prioritizing flag", flagParamName, envParamName)
			return
		} else if *flagParam == "" && envParam != "" {
			*flagParam = envParam
		}
	}

	isAuthFlagParamEmpty := clientID == "" && issuer == "" && audience == ""

	setParam(&udashAPIEndpoint, envVariableAPIURL, "API URL", DefaultEnvVariableAPIURL)
	setParam(&accessToken, envVariableAccessToken, "api access token", DefaultEnvVariableAccessToken)
	setParam(&udashEndpoint, envVariableURL, "URL", DefaultEnvVariableURL)

	if isAuthFlagParamEmpty {
		logrus.Debugf("No authentication parameters provided, skipping authorization")

		err := updateConfigFile(authData{
			URL:   udashEndpoint,
			API:   udashAPIEndpoint,
			Token: envVariableAccessToken,
		})

		if err != nil {
			return fmt.Errorf("update Updatecli config file: %w", err)
		}

		return nil

	} else if udashEndpoint == "" {
		return fmt.Errorf("service URL is required")
	}

	port, err := getAvailablePort()
	if err != nil {
		return fmt.Errorf("get available port: %w", err)
	}

	err = authorizeUser(
		udashEndpoint,
		clientID,
		issuer,
		audience,
		fmt.Sprintf("http://localhost:%s", port),
		accessToken,
	)

	if err != nil {
		return err
	}

	return nil
}
