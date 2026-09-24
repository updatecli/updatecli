package app

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jferrl/go-githubauth"
	"golang.org/x/oauth2"
)

const (
	// DefaultExpirationTime defines the default expiration time for a GitHub App token
	// We set it to 3600 seconds (1 hour) to be sure the token is valid for the entire duration of the updatecli run
	DefaultExpirationTime int64 = 3600
	// MinimumExpirationTime defines the minimum expiration time for a GitHub App token
	// We set it to 600 seconds (10 minutes) to be sure the token is valid for the entire duration of the updatecli run
	MinimumExpirationTime int64 = 600
)

// Spec defines the GitHub App credentials used to authenticate with the GitHub API.
type Spec struct {
	// "clientid" defines the GitHub App client ID.
	//
	ClientID string `yaml:",omitempty"`
	// "privatekey" defines the PEM encoded private key of the GitHub App.
	//
	// remark:
	//   * "privatekey" or "privatekeypath" is required.
	//   * when both are set, "privatekey" takes precedence.
	//   * prefer "privatekeypath", to keep sensitive information out of the manifest.
	//
	PrivateKey string `yaml:",omitempty"`
	// "privatekeypath" defines the path to a file holding the PEM encoded private key of the GitHub App.
	//
	// remark:
	//   * "privatekey" or "privatekeypath" is required.
	//   * when both are set, "privatekey" takes precedence.
	//   * setting the value from an environment variable keeps sensitive information out of the manifest.
	//
	// example:
	//   * privatekeypath: '{{ requiredEnv "GITHUB_APP_PRIVATE_KEY_PATH" }}'
	//
	PrivateKeyPath string `yaml:",omitempty"`
	// "installationid" defines the GitHub App installation ID.
	//
	// remark:
	//   * the value must be an integer.
	//   * it is the ID shown in the URL https://github.com/settings/installation/<ID>
	//
	InstallationID string `yaml:",omitempty"`
	// "expirationtime" defines the lifetime of the GitHub App token, in seconds.
	//
	// default:
	//   3600
	//
	// remark:
	//   * the token is used during the whole Updatecli run, so it must stay valid until the run ends.
	//   * the minimum value is 600.
	//
	ExpirationTime string `yaml:",omitempty"`
}

// NewSpecFromEnv creates a new Spec from environment variables
// It returns nil if the required environment variables are not set or if the configuration is invalid
// Required environment variables:
// - UPDATECLI_GITHUB_APP_CLIENT_ID
// - UPDATECLI_GITHUB_APP_PRIVATE_KEY or UPDATECLI_GITHUB_APP_PRIVATE_KEY_PATH
// - UPDATECLI_GITHUB_APP_INSTALLATION_ID
func NewSpecFromEnv() *Spec {
	spec := &Spec{
		ClientID:       os.Getenv("UPDATECLI_GITHUB_APP_CLIENT_ID"),
		PrivateKey:     os.Getenv("UPDATECLI_GITHUB_APP_PRIVATE_KEY"),
		PrivateKeyPath: os.Getenv("UPDATECLI_GITHUB_APP_PRIVATE_KEY_PATH"),
		InstallationID: os.Getenv("UPDATECLI_GITHUB_APP_INSTALLATION_ID"),
		ExpirationTime: os.Getenv("UPDATECLI_GITHUB_APP_EXPIRATION_TIME"),
	}

	if err := spec.Validate(); err != nil {
		return nil
	}

	return spec
}

// getPrivateKey returns the GitHub App private key as a string
// It first checks if PrivateKey is set, if not it checks if PrivateKeyPath is set
// If neither are set, it returns an error
func (g Spec) getPrivateKey() (string, error) {
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
func (g Spec) getInstallationID() (int64, error) {
	return strconv.ParseInt(g.InstallationID, 10, 64)
}

// getExpirationTime returns the GitHub App token expiration time as an int64
// or 3600 if not set
func (g Spec) getExpirationTime() (int64, error) {
	if g.ExpirationTime == "" {
		return DefaultExpirationTime, nil
	}

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
func (g Spec) getExpirationTimeDuration() (time.Duration, error) {
	expirationTime, err := g.getExpirationTime()
	if err != nil {
		return 0, err
	}
	return time.Duration(expirationTime) * time.Second, nil
}

// Validate validates the GitHub App configuration
func (g *Spec) Validate() error {
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
		errMsg := "Github App configuration is invalid:"
		for i := range errs {
			errMsg = fmt.Sprintf("%s\n - %s", errMsg, errs[i].Error())
		}
		return errors.New(errMsg)
	}

	return nil
}

// Getoauth2TokenSource returns an oauth2.TokenSource to authenticate with GitHub using a GitHub App
func (g *Spec) Getoauth2TokenSource() (oauth2.TokenSource, error) {
	privateKey, err := g.getPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("retrieving GitHub App private key: %w", err)
	}
	clientID := g.ClientID
	installationID, err := g.getInstallationID()
	if err != nil {
		return nil, fmt.Errorf("invalid GitHub App installation ID: %w", err)
	}

	expirationTimeDuration, err := g.getExpirationTimeDuration()
	if err != nil {
		return nil, fmt.Errorf("invalid GitHub App expiration time: %w", err)
	}

	appTokenSource, err := githubauth.NewApplicationTokenSource(
		clientID,
		[]byte(privateKey),
		githubauth.WithApplicationTokenExpiration(expirationTimeDuration),
	)
	if err != nil {
		return nil, fmt.Errorf("creating GitHub App token source: %w", err)
	}

	tokenSource := githubauth.NewInstallationTokenSource(installationID, appTokenSource)

	return tokenSource, nil
}
