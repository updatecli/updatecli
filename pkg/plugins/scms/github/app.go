package github

import (
	"errors"
	"os"
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
	InstallationID int64 `yaml:",omitempty"`
}

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
