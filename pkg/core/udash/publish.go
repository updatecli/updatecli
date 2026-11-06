package udash

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/version"
)

var (
	// ErrNoUdashAPIURL is returned if we couldn't find an Updatecli report API
	ErrNoUdashAPIURL error = fmt.Errorf("no Updatecli API defined")
	// ErrAPINotFound is returned when no Udash API is served at the configured API URL,
	// typically because it predates the versioned API
	ErrAPINotFound error = errors.New("udash API not found")
)

// Publish publish a pipeline report to the updatecli api
func Publish(r *reports.Report) error {

	log.Infof("Publishing report to Udash")

	// setDefaultParam sets the default value for a parameter
	setDefaultParam := func(envParam *string, configParam, envParamName, configParamName string) {
		if *envParam != "" && configParam != "" {
			log.Debugf("%s provided via environment variable %q supersede value %q from %q in config file",
				*envParam,
				envParamName,
				configParamName,
				configParam)
			return
		} else if *envParam == "" && configParam != "" {
			*envParam = configParam
		}
	}

	envUdashURLString, envUdashApiURLString, envUdashToken := getConfigFromEnv()

	configUdashURLString, configUdashApiURLString, configUdashToken, err := getConfigFromFile("")
	if err != nil {
		log.Debugf("get Udash config from file: %s", err)
	}

	setDefaultParam(&envUdashApiURLString, configUdashApiURLString, DefaultEnvVariableAPIURL, "api")
	setDefaultParam(&envUdashURLString, configUdashURLString, DefaultEnvVariableURL, "url")
	setDefaultParam(&envUdashToken, configUdashToken, DefaultEnvVariableAccessToken, "token")

	if envUdashApiURLString == "" {
		return ErrNoUdashAPIURL
	}

	reportApiURL, err := url.Parse(envUdashApiURLString)
	if err != nil {
		return fmt.Errorf("parsing report API URL: %w", err)
	}

	reportURL, err := url.Parse(envUdashURLString)
	if err != nil {
		return fmt.Errorf("parsing report URL: %w", err)
	}

	r.UpdatecliVersion = version.Version

	jsonBody, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling json: %w", err)
	}

	u := reportApiURL.JoinPath("pipeline", "reports")

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}

	if envUdashToken != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", envUdashToken))
	}

	res, err := httpclient.NewRetryClient().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response from %s: %w", u.String(), err)
	}

	if res.StatusCode >= 400 {
		log.Debugf("response from %s:\n%s", u.String(), string(data))
		return responseError(res.Status, res.StatusCode, data, envUdashURLString, envUdashApiURLString)
	}

	d := struct {
		ReportID string
		Message  string
	}{}

	if err := json.Unmarshal(data, &d); err != nil {
		return fmt.Errorf("decoding response from %s: %w", u.String(), err)
	}

	// Without a front URL or a report ID, the link would be a broken relative path.
	if envUdashURLString != "" && d.ReportID != "" {
		r.ReportURL = reportURL.JoinPath("pipeline", "reports", d.ReportID).String()
	}

	return nil
}

// maxErrorBodyLength bounds how much of a Udash error response ends up in an error message.
const maxErrorBodyLength = 512

// responseError builds the error returned when Udash refuses a report.
func responseError(status string, statusCode int, body []byte, udashURL, apiURL string) error {
	msg := strings.TrimSpace(string(body))
	if len(msg) > maxErrorBodyLength {
		msg = msg[:maxErrorBodyLength] + "..."
	}

	err := fmt.Errorf("udash responded %s", status)
	if msg != "" {
		err = fmt.Errorf("udash responded %s: %s", status, msg)
	}

	loginURL := udashURL
	if loginURL == "" {
		loginURL = "<url>"
	}

	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		return fmt.Errorf("%w, run `updatecli udash login %s` to store a valid token", err, loginURL)
	case http.StatusNotFound:
		// Publishing only ever creates a report, so a 404 means the API is not served
		// at that base, typically an API URL stored before Udash versioned its API.
		return fmt.Errorf("%w: %w at %s, Udash serves it under <url>%s, run `updatecli udash login %s` again or set %s",
			err, ErrAPINotFound, apiURL, DefaultAPIPath, loginURL, DefaultEnvVariableAPIURL)
	}

	return err
}
