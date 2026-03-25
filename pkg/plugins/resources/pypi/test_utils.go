package pypi

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/httpclient"
)

const existingPackageData = `{
  "info": {
    "name": "requests",
    "version": "2.31.0",
    "project_urls": {
      "Source": "https://github.com/psf/requests",
      "Changelog": "https://github.com/psf/requests/blob/main/HISTORY.md"
    }
  },
  "releases": {
    "2.28.0": [{"yanked": false}],
    "2.29.0": [{"yanked": false}],
    "2.30.0": [{"yanked": true}],
    "2.31.0": [{"yanked": false}]
  }
}`

// yankedPackageData has only yanked releases except one early version.
const yankedPackageData = `{
  "info": {
    "name": "requests",
    "version": "2.28.0",
    "project_urls": {
      "Source": "https://github.com/psf/requests"
    }
  },
  "releases": {
    "2.28.0": [{"yanked": false}],
    "2.29.0": [{"yanked": true}],
    "2.30.0": [{"yanked": true}],
    "2.31.0": [{"yanked": true}]
  }
}`

const preReleasePackageData = `{
  "info": {
    "name": "testpkg",
    "version": "1.0b2",
    "project_urls": {}
  },
  "releases": {
    "1.0a1": [{"yanked": false}],
    "1.0b2": [{"yanked": false}],
    "0.9.0": [{"yanked": false}]
  }
}`

const nonExistingPackageData = `{"message": "Not Found"}`

// GetMockClient returns a MockClient that validates the URL prefix and Bearer token,
// then serves the provided body and status code.
func GetMockClient(baseURL, mockedToken, mockedBody string, mockedHTTPStatusCode int) *httpclient.MockClient {
	return &httpclient.MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			var statusCode int
			var httpError error
			var body string

			if !strings.HasPrefix(req.URL.String(), baseURL) {
				statusCode = 404
				httpError = errors.New("not found")
			} else if mockedToken != "" && req.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", mockedToken) {
				statusCode = 401
				httpError = errors.New("unauthorized")
			} else {
				body = mockedBody
				statusCode = mockedHTTPStatusCode
			}

			return &http.Response{
				StatusCode: statusCode,
				Body:       io.NopCloser(strings.NewReader(body)),
			}, httpError
		},
	}
}
