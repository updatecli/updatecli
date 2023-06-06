package reports

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/auth"
)

var (
	// ErrNoBearerToken is returned if we couldn't find a token in the local updatecli configuration file
	ErrNoBearerToken error = fmt.Errorf("no bearer token found")
	// ErrNoAuthAudience is returned if we couldn't find an Oauth audience
	ErrNoOAuthAudience error = fmt.Errorf("no Oauth audience defined")
)

// Publish publish a pipeline report to the updatecli api
func (r Report) Publish() error {

	if auth.OauthAudience == "" {
		return ErrNoOAuthAudience
	}

	bearerToken, err := auth.Token(auth.OauthAudience)
	if err != nil {
		return fmt.Errorf("get bearer token: %w", err)
	}

	if bearerToken == "" {
		return ErrNoBearerToken
	}

	jsonBody, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshaling json: %w", err)
	}

	bodyReader := bytes.NewReader(jsonBody)

	url, err := url.JoinPath(auth.OauthAudience, "api", "pipeline", "reports")
	if err != nil {
		return fmt.Errorf("generating URL: %w", err)
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", bearerToken))

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		body, err := httputil.DumpResponse(resp, false)
		if err != nil {
			return err
		}
		logrus.Debugf("\n%v\n", string(body))
	}

	return nil
}

func (r Reports) Publish() error {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Report"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Report")+1))

	for _, report := range r {
		err := report.Publish()
		if err != nil &&
			!errors.Is(err, ErrNoBearerToken) &&
			!errors.Is(err, ErrNoOAuthAudience) {
			logrus.Debugf("publish report: %s", err)
		}
	}
	return nil
}
