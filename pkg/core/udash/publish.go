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
	"github.com/updatecli/updatecli/pkg/core/reports"
)

var (
	// ErrNoUdashAPIURL is returned if we couldn't find an Updatecli report API
	ErrNoUdashAPIURL error = fmt.Errorf("no Updatecli API defined")
)

// Publish publish a pipeline report to the updatecli api
func Publish(r *reports.Report) error {
	err := r.UpdateID()
	if err != nil {
		return fmt.Errorf("generating report IDs: %w", err)
	}

	reportURLString, reportApiURLString, bearerToken, err := Token("")
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

	jsonBody, err := json.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshaling json: %w", err)
	}

	bodyReader := bytes.NewReader(jsonBody)

	u := reportApiURL.JoinPath("pipeline", "reports")

	client := &http.Client{}

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
