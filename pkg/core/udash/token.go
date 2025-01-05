package udash

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// getConfigFromFile return the Udash configuration from the configuration file
func getConfigFromFile(audience string) (URL string, ApiURL string, Token string, err error) {
	if Audience != "" {
		audience = Audience
	}

	data, err := readConfigFile()
	if err != nil {
		return "", "", "", err
	}

	switch audience {
	case "":
		authdata, ok := data.Auths[data.Default]
		if ok {
			return authdata.URL, authdata.API, authdata.Token, nil
		}
		return "", "", "", fmt.Errorf("no default token found")
	default:
		authdata, ok := data.Auths[sanitizeTokenID(audience)]
		if ok {
			return authdata.URL, authdata.API, authdata.Token, nil
		}
	}

	return "", "", "", fmt.Errorf("token for domain %q not found", audience)
}

// getConfigFromEnv return the Udash configuration from environment variables
func getConfigFromEnv() (URL string, ApiURL string, Token string) {
	return os.Getenv(DefaultEnvVariableURL), os.Getenv(DefaultEnvVariableAPIURL), os.Getenv(DefaultEnvVariableAccessToken)
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
