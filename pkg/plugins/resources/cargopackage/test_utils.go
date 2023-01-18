package cargopackage

import (
	"errors"
	"fmt"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func CreateDummyIndex() (string, error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}
	index, err := os.Create(filepath.Join(dir, "config.json"))
	if err != nil {
		return "", err
	}
	defer index.Close()
	_, err = fmt.Fprintf(index, "{\"dl\":\"https://example.com\"}")
	if err != nil {
		return "", err
	}
	crateDir := filepath.Join(dir, "cr/at")
	err = os.MkdirAll(crateDir, 0750)
	if err != nil {
		return "", err
	}
	crateFile, err := os.Create(filepath.Join(crateDir, "crate-test"))
	if err != nil {
		return "", err
	}
	defer crateFile.Close()
	_, err = fmt.Fprintf(crateFile, "{\"name\":\"crate-test\",\"vers\":\"0.1.0\",\"deps\":[],\"features\":{},\"cksum\":\"b274d286f7a6aad5a7d5b5407e9db0098c94711fb3563bf2e32854a611edfb63\",\"yanked\":false,\"links\":null}\n")
	if err != nil {
		return "", err
	}
	_, err = fmt.Fprintf(crateFile, "{\"name\":\"crate-test\",\"vers\":\"0.2.0\",\"deps\":[],\"features\":{},\"cksum\":\"b274d286f7a6aad5a7d5b5407e9db0098c94711fb3563bf2e32854a611edfb63\",\"yanked\":false,\"links\":null}\n")
	if err != nil {
		return "", err
	}
	_, err = fmt.Fprintf(crateFile, "{\"name\":\"crate-test\",\"vers\":\"0.2.2\",\"deps\":[],\"features\":{},\"cksum\":\"b274d286f7a6aad5a7d5b5407e9db0098c94711fb3563bf2e32854a611edfb63\",\"yanked\":false,\"links\":null}\n")
	if err != nil {
		return "", err
	}
	_, err = fmt.Fprintf(crateFile, "{\"name\":\"crate-test\",\"vers\":\"0.2.3\",\"deps\":[],\"features\":{},\"cksum\":\"b274d286f7a6aad5a7d5b5407e9db0098c94711fb3563bf2e32854a611edfb63\",\"yanked\":true,\"links\":null}\n")
	if err != nil {
		return "", err
	}
	return dir, nil
}

const existingPackageData = `{
  "categories": [],
  "crate": {
    "badges": [],
    "categories": [],
    "created_at": "2023-01-12T16:51:06.647066+00:00",
    "description": "Test Package Crate",
    "documentation": null,
    "downloads": 44,
    "exact_match": false,
    "homepage": null,
    "id": "crate-test",
    "keywords": [],
    "links": {},
    "max_stable_version": "0.2.0",
    "max_version": "0.2.0",
    "name": "crate-test",
    "newest_version": "0.2.0",
    "recent_downloads": 44,
    "repository": "https://github.com/test/test",
    "updated_at": "2023-01-15T19:00:34.723908+00:00",
    "versions": [
      704063,
      701926
    ]
  },
  "keywords": [],
  "versions": [
    {
      "crate": "crate-test",
      "id": 704063,
      "num": "0.2.0",
      "yanked": false
    },
    {
      "crate": "crate-test",
      "id": 701926,
      "num": "0.1.0",
      "yanked": false
    }
  ]
}`
const existingPackageStatus = 200
const nonExistingPackageData = `{"errors":[{"detail":"Not Found"}]}`
const nonExistingPackageStatus = 404

func GetMockClient(baseUrl string, mockedToken string, mockedBody string, mockedHTTPStatusCode int, mockedHeaderFormat string) *httpclient.MockClient {

	return &httpclient.MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			var statusCode int
			var httpError error
			var body string
			if !strings.HasPrefix(req.URL.String(), baseUrl) {
				statusCode = 404
				httpError = errors.New("not found")
			} else if req.Header.Get("Authorization") != fmt.Sprintf(mockedHeaderFormat, mockedToken) {
				statusCode = 401
				httpError = errors.New("unauthorized")
			} else {
				body = mockedBody
				statusCode = mockedHTTPStatusCode
			}
			return &http.Response{
				StatusCode: statusCode,
				Body:       ioutil.NopCloser(strings.NewReader(body)),
			}, httpError
		},
	}
}
