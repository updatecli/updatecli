package github

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	// DefaultExpirationTime defines the default expiration time for a GitHub App token
	// We set it to 10800 seconds (3 hours) to be sure the token is valid for the entire duration of the updatecli run
	DefaultExpirationTime int64 = 10800
	// MinimumExpirationTime defines the minimum expiration time for a GitHub App token
	// We set it to 600 seconds (10 minutes) to be sure the token is valid for the entire duration of the updatecli run
	MinimumExpirationTime int64 = 600
)

// GitHubAppSpec defines the specification to authenticate with a GitHub App
type GitHubAppSpec struct {
	// ClientID represents the GitHub App client ID
	ClientID string `yaml:",omitempty"`
	// PrivateKey represents a PEM encoded private key
	// It is recommended to use PrivateKeyPath instead of PrivateKey
	// to avoid putting sensitive information in the configuration file
	// If both PrivateKey and PrivateKeyPath are set, PrivateKey takes precedence
	PrivateKey string `yaml:",omitempty"`
	// PrivateKeyPath represents the path to a PEM encoded private key
	// If both PrivateKey and PrivateKeyPath are set, PrivateKey takes precedence
	// It is recommended to use an environment variable to set the PrivateKeyPath value
	// e.g. PrivateKeyPath: {{ requiredEnv "GITHUB_APP_PRIVATE_KEY_PATH" }}
	// to avoid putting sensitive information in the configuration file
	PrivateKeyPath string `yaml:",omitempty"`
	// InstallationID represents the GitHub App installation ID
	// It is the same ID that you can find in the GitHub endpoint:
	// https://github.com/settings/installation/<ID>
	InstallationID string `yaml:",omitempty"`
	// Expiration represents the token expiration time in seconds
	// The token is used during the entire execution of updatecli
	// and should be valid for the entire duration of the run
	// The minimum value is 600 seconds (10 minutes)
	//
	// Default: 10800 (3 hours)
	ExpirationTime string `yaml:",omitempty"`
}

// getPrivateKey returns the GitHub App private key as a string
// It first checks if PrivateKey is set, if not it checks if PrivateKeyPath is set
// If neither are set, it returns an error
func (g GitHubAppSpec) getPrivateKey() (string, error) {
	if g.PrivateKey != "" {
		return g.PrivateKey, nil
	}
	if g.PrivateKeyPath != "" {
		privateKeyBytes, err := os.ReadFile(g.PrivateKeyPath)
		if err != nil {
			return "", err
		}
		return string(privateKeyBytes), nil
	}
	return "", errors.New("no private key or private key path provided")
}

// getInstallationID returns the GitHub App installation ID as an int64
func (g GitHubAppSpec) getInstallationID() (int64, error) {
	return strconv.ParseInt(g.InstallationID, 10, 64)
}

// getExpirationTime returns the GitHub App token expiration time as an int64
// or 6000 if not set
func (g GitHubAppSpec) getExpirationTime() (int64, error) {
	expirationTime, err := strconv.ParseInt(g.ExpirationTime, 10, 64)
	if err != nil {
		return 0, err
	}

	if expirationTime == 0 {
		return DefaultExpirationTime, nil
	}

	return expirationTime, nil
}

// getExpirationTimeDuration returns the GitHub App token expiration time as a time.Duration
func (g GitHubAppSpec) getExpirationTimeDuration() (time.Duration, error) {
	expirationTime, err := g.getExpirationTime()
	if err != nil {
		return 0, err
	}
	return time.Duration(expirationTime) * time.Second, nil
}

// Validate validates the GitHub App configuration

func (g *GitHubAppSpec) Validate() error {

	var errs []error

	if g.ClientID == "" {
		errs = append(errs, errors.New("github app client id is not set"))
	}

	if g.InstallationID == "" {
		errs = append(errs, errors.New("github app installation id is not set"))
	}

	if g.PrivateKey == "" && g.PrivateKeyPath == "" {
		errs = append(errs, errors.New("github app private key or private key path is not set"))
	}

	if _, err := g.getInstallationID(); err != nil {
		errs = append(errs, errors.New("github app installation id is not a valid integer"))
	}

	expirationTime, err := g.getExpirationTime()
	if err != nil {
		errs = append(errs, errors.New("github app token expiration time is not a valid integer"))
	}

	if expirationTime < MinimumExpirationTime {
		errs = append(errs, fmt.Errorf("github app token expiration time should be at least %d seconds", MinimumExpirationTime))
	}

	if len(errs) > 0 {
		logrus.Errorf("Github App configuration is invalid:")
		for i := range errs {
			logrus.Errorf(" - %s", errs[i].Error())
		}
		return errors.New("github app configuration is invalid")
	}

	return nil

}
