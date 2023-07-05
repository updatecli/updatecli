package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Token return the token for a specific auth domain
func Token(audience string) (string, error) {
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

	encodedAudience := base64.StdEncoding.EncodeToString([]byte(sanitizeTokenID(audience)))

	authdata, ok := data.Auths[strings.ToLower(encodedAudience)]
	if ok {
		return authdata.Auth, nil
	}

	return "", fmt.Errorf("token for domain %q not found", audience)
}

// getAccessToken trades the authorization code retrieved from the first OAuth2 log for an access token
func getAccessToken(issuer, clientID, codeVerifier, authorizationCode, callbackURL string) (string, error) {
	u, err := url.Parse(issuer)
	if err != nil {
		return "", err
	}

	u = u.JoinPath("oauth", "token")

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("code_verifier", codeVerifier)
	data.Set("code", authorizationCode)
	data.Set("scope", "offline_access")
	data.Set("redirect_uri", callbackURL)

	payload := strings.NewReader(data.Encode())

	// create the request and execute it
	req, _ := http.NewRequest("POST", u.String(), payload)
	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Printf("updatecli: HTTP error: %s", err)
		return "", err
	}

	// process the response
	defer res.Body.Close()
	var responseData map[string]interface{}
	body, _ := io.ReadAll(res.Body)

	// unmarshal the json into a string map
	err = json.Unmarshal(body, &responseData)
	if err != nil {
		logrus.Printf("updatecli: JSON error: %s", err)
		return "", err
	}

	// retrieve the access token out of the map, and return to caller
	accessToken := responseData["access_token"].(string)
	return accessToken, nil
}
