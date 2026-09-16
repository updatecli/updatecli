package osv

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// maxPages bounds the pagination loop in case the API keeps returning a page token.
const maxPages = 100

// osvPackage identifies a package in OSV queries and records.
type osvPackage struct {
	Ecosystem string `json:"ecosystem,omitempty"`
	Name      string `json:"name,omitempty"`
	Purl      string `json:"purl,omitempty"`
}

// queryRequest is the body of POST /v1/query.
type queryRequest struct {
	Package   osvPackage `json:"package"`
	Version   string     `json:"version,omitempty"`
	PageToken string     `json:"page_token,omitempty"`
}

// queryResponse is the body returned by POST /v1/query.
type queryResponse struct {
	Vulns         []vulnerability `json:"vulns"`
	NextPageToken string          `json:"next_page_token"`
}

// vulnerability holds the subset of an OSV record used by the plugin.
type vulnerability struct {
	ID        string     `json:"id"`
	Aliases   []string   `json:"aliases"`
	Summary   string     `json:"summary"`
	Withdrawn string     `json:"withdrawn"`
	Affected  []affected `json:"affected"`
	// DatabaseSpecific is free-form, its content depends on the database publishing the record.
	DatabaseSpecific map[string]any `json:"database_specific"`
}

type affected struct {
	Package osvPackage `json:"package"`
	Ranges  []osvRange `json:"ranges"`
}

type osvRange struct {
	Type   string     `json:"type"`
	Events []osvEvent `json:"events"`
}

type osvEvent struct {
	Introduced   string `json:"introduced"`
	Fixed        string `json:"fixed"`
	LastAffected string `json:"last_affected"`
	Limit        string `json:"limit"`
}

// severity returns the database specific severity, as published by GitHub advisories.
func (v vulnerability) severity() string {
	severity, _ := v.DatabaseSpecific["severity"].(string)
	return severity
}

// query returns every OSV record affecting the given version of the package.
func (o *Osv) query(ctx context.Context, version string) ([]vulnerability, error) {
	request := queryRequest{
		Package: osvPackage{
			Ecosystem: o.spec.Ecosystem,
			Name:      o.spec.Name,
			Purl:      o.spec.Purl,
		},
		Version: version,
	}

	var vulns []vulnerability
	for range maxPages {
		response, err := o.queryPage(ctx, request)
		if err != nil {
			return nil, err
		}

		vulns = append(vulns, response.Vulns...)

		if response.NextPageToken == "" {
			return vulns, nil
		}
		request.PageToken = response.NextPageToken
	}

	return nil, fmt.Errorf("OSV query for version %q of %s exceeded %d pages", version, o.packageLabel(), maxPages)
}

// queryPage sends a single POST /v1/query request.
func (o *Osv) queryPage(ctx context.Context, request queryRequest) (queryResponse, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return queryResponse{}, fmt.Errorf("marshaling OSV query: %w", err)
	}

	requestURL := o.spec.URL + "/v1/query"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, bytes.NewReader(requestBody))
	if err != nil {
		return queryResponse{}, fmt.Errorf("building request for %q: %w", requestURL, err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := o.webClient.Do(req)
	if err != nil {
		return queryResponse{}, fmt.Errorf("querying OSV for %s: %w", o.packageLabel(), err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return queryResponse{}, fmt.Errorf("reading OSV response body: %w", err)
	}

	if res.StatusCode >= 400 {
		return queryResponse{}, fmt.Errorf("OSV API returned HTTP %d for %s: %s", res.StatusCode, o.packageLabel(), strings.TrimSpace(string(body)))
	}

	var response queryResponse
	if err = json.Unmarshal(body, &response); err != nil {
		return queryResponse{}, fmt.Errorf("unmarshaling OSV response: %w", err)
	}

	return response, nil
}
