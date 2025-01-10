package udash

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/reports"
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

	err := r.UpdateID()
	if err != nil {
		return fmt.Errorf("generating report IDs: %w", err)
	}

	reportURLString, reportApiURLString, bearerToken := getConfigFromEnv()
	if reportApiURLString == "" {
		logrus.Debugf("no Udash API URL detected via environment variables, looking for a configuration file")
		localReportURLString, localReportApiURLString, localBearerToken, err := getConfigFromFile("")
		if err != nil {
			logrus.Debugf("retrieving service access token: %s", err)
		}

		setDefaultParam(&reportApiURLString, localReportApiURLString, DefaultEnvVariableAPIURL, "api")
		setDefaultParam(&reportURLString, localReportURLString, DefaultEnvVariableURL, "url")
		setDefaultParam(&bearerToken, localBearerToken, DefaultEnvVariableAccessToken, "token")
	}

	if reportApiURLString == "" {
		logrus.Infof("no Udash endpoint detected, skipping report publication")
		return nil
	}

	reportApiURL, err := url.Parse(reportApiURLString)
	if err != nil {
		return fmt.Errorf("parsing report API URL: %w", err)
	}

	reportURL, err := url.Parse(reportURLString)
	if err != nil {
		return fmt.Errorf("parsing report URL: %w", err)
	}

	jsonBody, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling json: %w", err)
	}

	bodyReader := bytes.NewReader(jsonBody)

	u := reportApiURL.JoinPath("pipeline", "reports")

	client := httpclient.NewRetryClient()

	req, err := http.NewRequest("POST", u.String(), bodyReader)
	if err != nil {
		return err
	}

	if bearerToken != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", bearerToken))
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode >= 400 {
		body, err := httputil.DumpResponse(res, false)
		if err != nil {
			return err
		}
		logrus.Debugf("\n%v\n", string(body))
	}

	defer res.Body.Close()
	if res.StatusCode >= 400 {
		body, err := httputil.DumpResponse(res, false)
		logrus.Debugf("\n%v\n", string(body))
		return err
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		logrus.Debugf("\n%v\n", string(data))
		return err
	}

	d := struct {
		ReportID string
		Message  string
	}{}

	err = json.Unmarshal(data, &d)
	if err != nil {
		logrus.Errorf("error unmarshalling json: %q", err)
		return err
	}

	r.ReportURL = reportURL.JoinPath("pipeline", "reports", d.ReportID).String()

	return nil
}
