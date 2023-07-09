package reports

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	// ErrNoReportAPIURL is returned if we couldn't find an Updatecli report API
	ErrNoReportAPIURL error = fmt.Errorf("no Updatecli API defined")

	// EnvReportURL defines the default environment variable use to define the updatecli report url
	EnvReportURL = "UPDATECLI_REPORT_URL"
	// EnvReportAPIURL defines the default environment variable use to define the updatecli report url
	EnvReportAPIURL = "UPDATECLI_REPORT_API_URL"
)

// Publish publish a pipeline report to the updatecli api
func (r *Report) Publish() error {
	logrus.Infof("\n\n%s\n", strings.ToTitle("Report"))
	logrus.Infof("%s\n\n", strings.Repeat("=", len("Report")+1))

	err := r.updateID()
	if err != nil {
		return fmt.Errorf("generating report IDs: %w", err)
	}

	reportURLString, reportApiURLString, bearerToken, err := auth.Token("")
	if err != nil {
		return fmt.Errorf("retrieving service access token: %w", err)
	}

	reportApiURL, err := url.Parse(reportApiURLString)
	if err != nil {
		return fmt.Errorf("parsing report API URL: %w", err)
	}

	reportURL, err := url.Parse(reportURLString)
	if err != nil {
		return fmt.Errorf("parsing report URL: %w", err)
	}

	if bearerToken == "" {
		return ErrNoBearerToken
	}

	jsonBody, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshaling json: %w", err)
	}

	bodyReader := bytes.NewReader(jsonBody)

	u := reportApiURL.JoinPath("pipeline", "reports")
	if err != nil {
		return fmt.Errorf("generating URL: %w", err)
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", u.String(), bodyReader)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", bearerToken))

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
	logrus.Printf("Report available on:\n\t * %q", r.ReportURL)

	return nil
}
