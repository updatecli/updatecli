package udash

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/reports"
	"github.com/updatecli/updatecli/pkg/core/version"
)

var (
	// ErrNoUdashAPIURL is returned if we couldn't find an Updatecli report API
	ErrNoUdashAPIURL error = fmt.Errorf("no Updatecli API defined")
)

// Publish publish a pipeline report to the updatecli api
func Publish(r *reports.Report) error {

	logrus.Infof("Publishing report to Udash")

	// setDefaultParam sets the default value for a parameter
	setDefaultParam := func(envParam *string, configParam, envParamName, configParamName string) {
		if *envParam != "" && configParam != "" {
			logrus.Debugf("%s provided via environment variable %q supersede value %q from %q in config file",
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
		logrus.Debugf("get Udash config from file: %s", err)
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
		logrus.Debugf("response from %s:\n%s", u.String(), string(data))
		return responseError(res.Status, res.StatusCode, data, envUdashURLString)
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
func responseError(status string, statusCode int, body []byte, udashURL string) error {
	msg := strings.TrimSpace(string(body))
	if len(msg) > maxErrorBodyLength {
		msg = msg[:maxErrorBodyLength] + "..."
	}

	err := fmt.Errorf("udash responded %s", status)
	if msg != "" {
		err = fmt.Errorf("udash responded %s: %s", status, msg)
	}

	if statusCode == http.StatusUnauthorized || statusCode == http.StatusForbidden {
		loginURL := udashURL
		if loginURL == "" {
			loginURL = "<url>"
		}
		return fmt.Errorf("%w, run `updatecli udash login %s` to store a valid token", err, loginURL)
	}

	return err
}
