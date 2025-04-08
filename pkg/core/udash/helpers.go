package udash

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

// setDefaultHTTPSScheme adds https:// to a URL if it doesn't already have a scheme
func setDefaultHTTPSScheme(URL string) string {
	if !strings.HasPrefix(URL, "http://") &&
		!strings.HasPrefix(URL, "https://") {
		URL = "https://" + URL
	}

	return URL
}

// sanitizeTokenID removes the scheme and trailing slash from a URL
func sanitizeTokenID(token string) string {
	token = strings.TrimPrefix(token, "https://")
	token = strings.TrimPrefix(token, "http://")
	token = strings.TrimSuffix(token, "/")

	token = strings.ToLower(token)
	return token
}

// initConfigFile creates Updatecli config directory
func initConfigFile() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		logrus.Errorln(err)
		return "", err
	}

	updatecliConfigDir := filepath.Join(userConfigDir, "updatecli")

	if _, err := os.Stat(updatecliConfigDir); errors.Is(err, fs.ErrNotExist) {
		err := os.MkdirAll(updatecliConfigDir, 0755)
		if err != nil {
			return "", err
		}
	}

	return filepath.Join(updatecliConfigDir, "udash.json"), nil
}

// IsConfigFile checks if the udash config file exists
func IsConfigFile() (string, bool) {
	configFile, err := initConfigFile()
	if err != nil {
		return configFile, false
	}

	if _, err := os.Stat(configFile); err != nil {
		return configFile, false
	}

	return configFile, true
}
