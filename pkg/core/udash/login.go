package udash

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/term"
)

// TokenPagePath is the Udash frontend path where a user manages their API tokens.
const TokenPagePath = "/profile/tokens"

// Login stores an Udash API token in the Updatecli configuration file.
//
// Udash issues the token itself, so it does not expire and Updatecli never has to
// talk to the identity provider: it only ever carries the token as a bearer.
func Login(udashEndpoint, udashAPIEndpoint, token string) error {

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

	setParam(&udashEndpoint, envVariableURL, "URL", DefaultEnvVariableURL)
	setParam(&udashAPIEndpoint, envVariableAPIURL, "API URL", DefaultEnvVariableAPIURL)
	setParam(&token, envVariableAccessToken, "api access token", DefaultEnvVariableAccessToken)

	if udashEndpoint == "" {
		return fmt.Errorf("service URL is required")
	}

	udashEndpoint = setDefaultHTTPSScheme(udashEndpoint)

	if udashAPIEndpoint == "" {
		udashAPIEndpoint = strings.TrimSuffix(udashEndpoint, "/") + "/api"
	}
	udashAPIEndpoint = setDefaultHTTPSScheme(udashAPIEndpoint)

	if token == "" {
		// A service running without authentication has no token to ask for, and
		// that is the default deployment.
		needed, err := requiresToken(udashAPIEndpoint)
		if err != nil {
			return fmt.Errorf("reaching %s: %w", udashAPIEndpoint, err)
		}

		if needed {
			token, err = readToken(udashEndpoint)
			if err != nil {
				return err
			}

			if token == "" {
				return fmt.Errorf("no token provided")
			}
		} else {
			logrus.Debugf("%s does not require authentication, storing the endpoint only", udashEndpoint)
		}
	}

	// Fail here rather than silently later on, when a pipeline tries to publish.
	identity, err := whoami(udashAPIEndpoint, token)
	if err != nil {
		return fmt.Errorf("validating token: %w", err)
	}

	if err := updateConfigFile(authData{
		URL:   udashEndpoint,
		API:   udashAPIEndpoint,
		Token: token,
	}); err != nil {
		return fmt.Errorf("update Updatecli config file: %w", err)
	}

	switch {
	case identity.Name != "":
		logrus.Printf("Successfully logged in to %s as %s.", udashEndpoint, identity.Name)
	case identity.Subject != "":
		logrus.Printf("Successfully logged in to %s as %s.", udashEndpoint, identity.Subject)
	default:
		logrus.Printf("Successfully logged in to %s.", udashEndpoint)
	}

	return nil
}

// readToken obtains a token from the user, either by prompting for one on an
// interactive terminal or by reading standard input when it is piped in.
func readToken(udashEndpoint string) (string, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		// Support `echo $TOKEN | updatecli udash login <url>`
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading token from stdin: %w", err)
		}
		return strings.TrimSpace(string(data)), nil
	}

	tokenPageURL := strings.TrimSuffix(udashEndpoint, "/") + TokenPagePath

	logrus.Printf("Create an API token at %s", tokenPageURL)
	if err := open.Start(tokenPageURL); err != nil {
		// Not fatal: the URL was printed above, the user can open it themselves.
		logrus.Debugf("can't open browser to URL %s: %s", tokenPageURL, err)
	}

	fmt.Print("Paste your Udash API token: ")
	data, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("reading token: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}
