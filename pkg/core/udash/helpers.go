package udash

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

func setDefaultHTTPSScheme(URL string) string {
	if !strings.HasPrefix(URL, "http://") &&
		!strings.HasPrefix(URL, "https://") {
		URL = "https://" + URL
	}

	return URL
}

func sanitizeTokenID(token string) string {
	token = strings.TrimPrefix(token, "https://")
	token = strings.TrimPrefix(token, "http://")
	token = strings.TrimSuffix(token, "/")

	// . are used by viper to split the key which is not compatible with dots used in URL
	token = strings.ReplaceAll(token, ".", "_")
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

	if _, err := os.Stat(updatecliConfigDir); os.IsNotExist(err) {
		err := os.MkdirAll(updatecliConfigDir, 0755)
		if err != nil {
			return "", err
		}
	}

	return filepath.Join(updatecliConfigDir, "config.json"), nil
}
