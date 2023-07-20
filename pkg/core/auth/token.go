package auth

import (
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

var (
	// DefaultEnvVariableAPIURL defines the default environment variable used to retrieve API url
	DefaultEnvVariableAPIURL string = "UPDATECLI_API_URL"
	// DefaultEnvVariableToken defines the default environment variable used to retrieve API access token
	DefaultEnvVariableToken string = "UPDATECLI_TOKEN"
	// DefaultEnvVariableURL defines the default environment variable used to retrieve url
	DefaultEnvVariableURL string = "UPDATECLI_URL"
)

// Token return the token for a specific auth domain
func Token(audience string) (URL string, ApiURL string, Token string, err error) {
	/*
		Exit early if the environment variable "UPDATECLI_TOKEN" contains a value.
	*/

	if Audience != "" {
		audience = Audience
	}

	if token := os.Getenv(DefaultEnvVariableToken); token != "" {
		logrus.Debugf("Token detect via env variable %q", DefaultEnvVariableToken)

		url := os.Getenv(DefaultEnvVariableURL)
		api := os.Getenv(DefaultEnvVariableAPIURL)

		if url == "" {
			logrus.Warningf("environment variable %q detected but missing value for %q", DefaultEnvVariableToken, DefaultEnvVariableURL)
		}
		if api == "" {
			logrus.Warningf("environment variable %q detected but missing value for %q", DefaultEnvVariableToken, DefaultEnvVariableAPIURL)
		}

		if token != "" && api != "" && url != "" {
			return url, api, token, nil
		}
		logrus.Warningf("Due to previous warning message, ignoring environment variable")
	}

	configFile, err := initConfigFile()
	if err != nil {
		return "", "", "", err
	}

	if _, err := os.Stat(configFile); errors.Is(err, os.ErrNotExist) {
		return "", "", "", err
	}

	configContent, err := os.ReadFile(configFile)
	if err != nil {
		return "", "", "", err
	}

	type authData struct {
		// Token stores the access token
		Token string
		// Api stores the api URL
		Api string
		// URL stores the front URL
		URL string
	}

	data := struct {
		Auths   map[string]authData
		Default string
	}{}

	if err := json.Unmarshal(configContent, &data); err != nil {
		return "", "", "", err
	}

	switch audience {
	case "":
		authdata, ok := data.Auths[data.Default]
		if ok {
			return authdata.URL, authdata.Api, authdata.Token, nil
		}
		return "", "", "", fmt.Errorf("no default token found")
	default:
		authdata, ok := data.Auths[sanitizeTokenID(audience)]
		if ok {
			return authdata.URL, authdata.Api, authdata.Token, nil
		}
	}

	return "", "", "", fmt.Errorf("token for domain %q not found", audience)
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
