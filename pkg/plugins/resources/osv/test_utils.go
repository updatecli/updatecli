package osv

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/httpclient"
)

var (
	// ghsaSandbox and pysecSandbox are the same vulnerability published by two databases.
	ghsaSandbox  = record("GHSA-cpwx-vrp4-4pq7", "MODERATE", []string{"CVE-2025-27516", "PYSEC-2026-1471"}, jinja2Affected("3.1.6"))
	pysecSandbox = record("PYSEC-2026-1471", "", []string{"CVE-2025-27516", "GHSA-cpwx-vrp4-4pq7"}, jinja2Affected("3.1.6"))
	// ghsaXSS and pysecXSS are the same vulnerability published by two databases.
	ghsaXSS  = record("GHSA-h75v-3vvj-5mfj", "MODERATE", []string{"CVE-2024-34064", "PYSEC-2026-1474"}, jinja2Affected("3.1.4"))
	pysecXSS = record("PYSEC-2026-1474", "", []string{"CVE-2024-34064", "GHSA-h75v-3vvj-5mfj"}, jinja2Affected("3.1.4"))
	ghsaHigh = record("GHSA-high-0000-0000", "HIGH", nil, jinja2Affected("3.1.3"))
	// pysecUnknown has no severity.
	pysecUnknown = record("PYSEC-2026-9999", "", nil, jinja2Affected("3.1.4"))
	withdrawn    = strings.Replace(record("GHSA-wdrn-0000-0000", "CRITICAL", nil, jinja2Affected("9.9.9")),
		`"summary"`, `"withdrawn": "2026-01-01T00:00:00Z", "summary"`, 1)
	// ghsaNew only affects versions released after the other vulnerabilities were fixed.
	ghsaNew = record("GHSA-new0-0000-0000", "MODERATE", nil, jinja2Affected("3.1.7"))
	// ghsaUnfixed has no fix published yet.
	ghsaUnfixed = record("GHSA-unfx-0000-0000", "HIGH", nil, jinja2Affected(""))

	// jinja2At312 holds 7 records for 4 vulnerabilities, once aliases are merged and the withdrawn record is dropped.
	jinja2At312 = response("", ghsaSandbox, pysecSandbox, ghsaXSS, pysecXSS, ghsaHigh, pysecUnknown, withdrawn)

	// xnetAt020 lists the Go standard library next to golang.org/x/net, as Go records do.
	xnetAt020 = response("", record("GO-2024-2687", "", []string{"CVE-2023-45288", "GHSA-4v7x-pqxf-cx7m"},
		affectedEntry("Go", "golang.org/x/net", "pkg:golang/golang.org/x/net", "SEMVER", "0.23.0"),
		affectedEntry("Go", "stdlib", "pkg:golang/stdlib", "SEMVER", "1.22.2"),
	))
)

// record returns an OSV record shaped like the api.osv.dev responses,
// including fields the plugin does not decode.
func record(id, severity string, aliases []string, affected ...string) string {
	databaseSpecific := `{"source": "https://example.com/advisory.json"}`
	if severity != "" {
		databaseSpecific = fmt.Sprintf(`{"cwe_ids": ["CWE-1336"], "github_reviewed": true, "severity": %q}`, severity)
	}

	aliasesJSON, _ := json.Marshal(aliases)

	return fmt.Sprintf(`{
  "id": %q,
  "summary": "Summary of %s",
  "modified": "2026-08-01T00:00:00Z",
  "published": "2025-03-05T00:00:00Z",
  "schema_version": "1.7.3",
  "aliases": %s,
  "severity": [{"type": "CVSS_V3", "score": "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:N/A:N"}],
  "affected": [%s],
  "references": [{"type": "WEB", "url": "https://example.com/%s"}],
  "database_specific": %s
}`, id, id, aliasesJSON, strings.Join(affected, ","), id, databaseSpecific)
}

// affectedEntry returns an affected entry whose single range is fixed in the given version.
func affectedEntry(ecosystem, name, purl, rangeType, fixed string) string {
	events := `{"introduced": "0"}`
	if fixed != "" {
		events += fmt.Sprintf(`, {"fixed": %q}`, fixed)
	}
	return fmt.Sprintf(`{"package": {"ecosystem": %q, "name": %q, "purl": %q}, "ranges": [{"type": %q, "events": [%s]}], "database_specific": {"source": "https://example.com"}}`,
		ecosystem, name, purl, rangeType, events)
}

func jinja2Affected(fixed string) string {
	return affectedEntry("PyPI", "jinja2", "pkg:pypi/jinja2", "ECOSYSTEM", fixed)
}

// response returns a POST /v1/query response body.
func response(nextPageToken string, records ...string) string {
	if nextPageToken == "" {
		return fmt.Sprintf(`{"vulns": [%s]}`, strings.Join(records, ","))
	}
	return fmt.Sprintf(`{"vulns": [%s], "next_page_token": %q}`, strings.Join(records, ","), nextPageToken)
}

// mockOSV serves POST /v1/query responses keyed by the queried version and page token.
type mockOSV struct {
	// responses maps "version|page_token" to a response body, unknown keys return no vulnerabilities.
	responses  map[string]string
	statusCode int
	requests   []queryRequest
}

func (m *mockOSV) client() *httpclient.MockClient {
	return &httpclient.MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			if req.Method != http.MethodPost || req.URL.String() != osvDefaultURL+"/v1/query" {
				return &http.Response{StatusCode: http.StatusNotFound, Body: io.NopCloser(strings.NewReader(""))}, nil
			}

			var query queryRequest
			if err := json.NewDecoder(req.Body).Decode(&query); err != nil {
				return nil, err
			}
			m.requests = append(m.requests, query)

			statusCode := m.statusCode
			if statusCode == 0 {
				statusCode = http.StatusOK
			}

			body, ok := m.responses[query.Version+"|"+query.PageToken]
			if !ok {
				body = "{}"
			}

			return &http.Response{StatusCode: statusCode, Body: io.NopCloser(strings.NewReader(body))}, nil
		},
	}
}

// requestedVersions returns the versions queried, in order.
func (m *mockOSV) requestedVersions() []string {
	versions := make([]string, 0, len(m.requests))
	for _, request := range m.requests {
		versions = append(versions, request.Version)
	}
	return versions
}
