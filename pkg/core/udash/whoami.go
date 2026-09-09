package udash

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/updatecli/updatecli/pkg/core/httpclient"
)

// whoamiResponse describes the identity behind a token, as reported by Udash.
type whoamiResponse struct {
	// Subject is the identity provider subject owning the token
	Subject string `json:"subject,omitempty"`
	// Name is a human readable name for that identity
	Name string `json:"name,omitempty"`
	// Permission is the Udash permission granted to that identity
	Permission string `json:"permission,omitempty"`
	// TokenName is the name given to the token, when authenticating with one
	TokenName string `json:"tokenName,omitempty"`
	// Scopes lists what the token is allowed to do, when authenticating with one
	Scopes []string `json:"scopes,omitempty"`
}

// callWhoami asks a Udash API who the given token belongs to.
//
// An empty token is sent without an Authorization header, which is how the caller
// finds out whether the service wants one at all.
func callWhoami(apiURL, token string) (*whoamiResponse, int, error) {
	u, err := url.Parse(setDefaultHTTPSScheme(apiURL))
	if err != nil {
		return nil, 0, fmt.Errorf("parsing API URL %q: %w", apiURL, err)
	}

	req, err := http.NewRequest("GET", u.JoinPath("whoami").String(), nil)
	if err != nil {
		return nil, 0, err
	}
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}

	resp, err := httpclient.NewRetryClient().Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, resp.StatusCode, nil
	}

	data := whoamiResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("decoding response from %s: %w", u.String(), err)
	}

	return &data, resp.StatusCode, nil
}

// requiresToken reports whether a Udash service authenticates its callers.
//
// The endpoint only exists when a mode is configured, so its absence means the
// service is open and Updatecli has nothing to log in with. Without this check
// the default deployment, which runs without authentication, could not be
// registered at all.
func requiresToken(apiURL string) (bool, error) {
	_, status, err := callWhoami(apiURL, "")
	if err != nil {
		return false, err
	}

	return status != http.StatusNotFound, nil
}

// whoami validates a token against a Udash API and returns the identity behind it.
func whoami(apiURL, token string) (*whoamiResponse, error) {
	identity, status, err := callWhoami(apiURL, token)
	if err != nil {
		return nil, err
	}

	switch {
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return nil, fmt.Errorf("token rejected by %s", apiURL)
	case status == http.StatusNotFound:
		// An older Udash has no /whoami endpoint: accept the token rather than
		// refusing to log in to a service which may otherwise work fine.
		return &whoamiResponse{}, nil
	case status >= 400:
		return nil, fmt.Errorf("unexpected response %d from %s", status, apiURL)
	}

	return identity, nil
}
