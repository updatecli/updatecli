package reports

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/auth"
)

var (
	// ErrNoBearerToken is returned if we couldn't find a token in the local updatecli configuration file
	ErrNoBearerToken error = fmt.Errorf("no bearer token found")
	// ErrNoReportAPIURL is returned if we couldn't find an Updatecli report API
	ErrNoReportAPIURL error = fmt.Errorf("no Updatecli API defined")
	// DefaultReportURL defines the default updatecli report url
	DefaultReportURL = "app.updatecli.io"
	// DefaultReportAPIURL defines the default updatecli report url
	DefaultReportAPIURL = "app.updatecli.io/api"

	// EnvReportURL defines the default environment variable use to define the updatecli report url
	EnvReportURL = "UPDATECLI_REPORT_URL"
	// EnvReportAPIURL defines the default environment variable use to define the updatecli report url
	EnvReportAPIURL = "UPDATECLI_REPORT_API_URL"
)

// Publish publish a pipeline report to the updatecli api
func (r *Report) Publish() error {

	err := r.updateID()
	if err != nil {
		return fmt.Errorf("generating report IDs: %w", err)
	}

	reportApiURL, err := parseURL(auth.OauthAudience, EnvReportAPIURL, DefaultReportAPIURL)
	if err != nil {
		return fmt.Errorf("parsing report API URL: %w", err)
	}
	reportURL, err := parseURL("", EnvReportURL, DefaultReportURL)
	if err != nil {
		return fmt.Errorf("parsing report URL: %w", err)
	}

	bearerToken, err := auth.Token(reportApiURL.String())
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

func (r *Reports) Publish() error {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Report"))
	logrus.Infof("%s\n\n", strings.Repeat("=", len("Report")+1))

	reports := *r
	for i, report := range reports {
		err := report.Publish()
		if err != nil &&
			!errors.Is(err, ErrNoBearerToken) &&
			!errors.Is(err, ErrNoReportAPIURL) {
			logrus.Errorf("publish report: %s", err)
		}
		reports[i] = report
	}
	return nil
}

// parseURL is a little helper function to parse an URL
func parseURL(param, env, def string) (url.URL, error) {
	var u *url.URL
	var err error

	if param != "" {
		u, err = url.Parse(auth.OauthAudience)
		if err != nil {
			return *u, fmt.Errorf("parsing URL: %w", err)
		}
	} else if os.Getenv(env) != "" {
		u, err = url.Parse(os.Getenv(env))
		if err != nil {
			return *u, fmt.Errorf("parsing URL: %w", err)
		}
	} else {
		u, err = url.Parse(def)
		if err != nil {
			return *u, fmt.Errorf("parsing URL: %w", err)
		}
	}
	return *u, nil
}
