package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// getOauthInfo queries the Udash website to retrieve Oauth configuration
func getOauthInfo(endpointURL string) (issuer string, audience string, clientID string, err error) {

	data := struct {
		Issuer   string `json:"OAUTH_DOMAIN,omitempty"`
		ClientID string `json:"OAUTH_CLIENTID,omitempty"`
		Audience string `json:"OAUTH_AUDIENCE,omitempty"`
	}{}

	endpointURL = setDefaultHTTPSScheme(endpointURL)

	URL, err := url.Parse(endpointURL)
	if err != nil {
		return "", "", "", fmt.Errorf("parsing endpoint URL %q: %v", endpointURL, err)
	}

	URL = URL.JoinPath("config.json")

	resp, err := http.Get(URL.String())
	if err != nil {
		return "", "", "", fmt.Errorf("cannot fetch URL %q: %v", URL.String(), err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", "", fmt.Errorf("unexpected http GET status: %s", resp.Status)
	}
	// We could check the resulting content type
	// here if desired.
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", "", "", fmt.Errorf("cannot decode JSON: %v", err)
	}
	return data.Issuer, data.Audience, data.ClientID, nil
}
