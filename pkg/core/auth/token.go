package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Token return the token for a specific auth domain
func Token(audience string) (string, error) {

	audience = sanitizeTokenID(audience)

	/*
		Exit early if the environment variable "UPDATECLI_API_TOKEN"
		contains a value.
	*/
	if token := os.Getenv("UPDATECLI_API_TOKEN"); token != "" {
		logrus.Debugln(`Environment variable UPDATECLI_API_TOKEN detected`)
		return token, nil
	}

	configFile, err := initConfigFile()
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
		if errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		return "", err
	}

	configContent, err := os.ReadFile(configFile)
	if err != nil {
		return "", err
	}

	type authData struct {
		Auth string
	}

	data := struct {
		Auths map[string]authData
	}{}

	if err := json.Unmarshal(configContent, &data); err != nil {
		return "", err
	}

	encodedAudience := make([]byte, base64.StdEncoding.EncodedLen(len(audience)))
	base64.StdEncoding.Encode(encodedAudience, []byte(audience))

	for i := range data.Auths {
		if strings.EqualFold(i, string(encodedAudience[:])) {
			return data.Auths[i].Auth, nil
		}
	}

	return "", fmt.Errorf("token for domain %q not found", audience)
}
