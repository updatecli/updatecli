package npm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/httpclient"
)

const existingPackageData = "{\"_id\":\"axios\",\"_rev\":\"779-b37ceeb27a03858a89a0226f7c554aaf\",\"name\":\"axios\",\"description\":\"Promise based HTTP client for the browser and node.js\",\"dist-tags\":{\"latest\":\"0.1.0\",\"next\":\"0.2.0\"},\"versions\":{\"0.1.0\":{\"name\":\"axios\",\"version\":\"0.1.0\",\"description\":\"Promise based XHR library\",\"main\":\"index.js\",\"scripts\":{\"test\":\"grunt test\",\"start\":\"node ./sandbox/index.js\"},\"repository\":{\"type\":\"git\",\"url\":\"https://github.com/mzabriskie/axios.git\"},\"keywords\":[\"xhr\",\"http\",\"ajax\",\"promise\"],\"author\":{\"name\":\"Matt Zabriskie\"},\"license\":\"MIT\",\"bugs\":{\"url\":\"https://github.com/mzabriskie/axios/issues\"},\"homepage\":\"https://github.com/mzabriskie/axios\",\"dependencies\":{\"es6-promise\":\"^1.0.0\"},\"devDependencies\":{\"grunt\":\"^0.4.5\",\"grunt-contrib-clean\":\"^0.6.0\",\"grunt-contrib-watch\":\"^0.6.1\",\"webpack\":\"^1.3.3-beta2\",\"webpack-dev-server\":\"^1.4.10\",\"grunt-webpack\":\"^1.0.8\",\"load-grunt-tasks\":\"^0.6.0\",\"karma\":\"^0.12.21\",\"karma-jasmine\":\"^0.1.5\",\"grunt-karma\":\"^0.8.3\",\"karma-phantomjs-launcher\":\"^0.1.4\",\"karma-jasmine-ajax\":\"^0.1.4\",\"grunt-update-json\":\"^0.1.3\",\"grunt-contrib-nodeunit\":\"^0.4.1\",\"grunt-banner\":\"^0.2.3\"},\"_id\":\"axios@0.1.0\",\"dist\":{\"shasum\":\"854e14f2999c2ef7fab058654fd995dd183688f2\",\"tarball\":\"https://registry.npmjs.org/axios/-/axios-0.1.0.tgz\",\"integrity\":\"sha512-hRPotWTy88LEsJ31RWEs2fmU7mV2YJs3Cw7Tk5XkKGtnT5NKOyIvPU+6qTWfwQFusxzChe8ozjay8r56wfpX8w==\",\"signatures\":[{\"keyid\":\"SHA256:jl3bwswu80PjjokCgh0o2w5c2U4LhQAE57gj9cz1kzA\",\"sig\":\"MEYCIQC/cOvHsV7UqLAet6WE89O4Ga3AUHgkqqoP0riLs6sgTAIhAIrePavu3Uw0T3vLyYMlfEI9bqENYjPzH5jGK8vYQVJK\"}]},\"_from\":\"./\",\"_npmVersion\":\"1.4.3\",\"_npmUser\":{\"name\":\"mzabriskie\",\"email\":\"mzabriskie@gmail.com\"},\"maintainers\":[{\"name\":\"mzabriskie\",\"email\":\"mzabriskie@gmail.com\"}],\"directories\":{},\"deprecated\":\"Critical security vulnerability fixed in v0.21.1. For more information, see https://github.com/axios/axios/pull/3410\"},\"0.2.0\":{\"name\":\"axios\",\"version\":\"0.2.0\",\"description\":\"Promise based HTTP client for the browser and node.js\",\"main\":\"index.js\",\"scripts\":{\"test\":\"grunt test\",\"start\":\"node ./sandbox/server.js\"},\"repository\":{\"type\":\"git\",\"url\":\"https://github.com/mzabriskie/axios.git\"},\"keywords\":[\"xhr\",\"http\",\"ajax\",\"promise\",\"node\"],\"author\":{\"name\":\"Matt Zabriskie\"},\"license\":\"MIT\",\"bugs\":{\"url\":\"https://github.com/mzabriskie/axios/issues\"},\"homepage\":\"https://github.com/mzabriskie/axios\",\"dependencies\":{\"es6-promise\":\"^1.0.0\"},\"devDependencies\":{\"grunt\":\"^0.4.5\",\"grunt-contrib-clean\":\"^0.6.0\",\"grunt-contrib-watch\":\"^0.6.1\",\"webpack\":\"^1.3.3-beta2\",\"webpack-dev-server\":\"^1.4.10\",\"grunt-webpack\":\"^1.0.8\",\"load-grunt-tasks\":\"^0.6.0\",\"karma\":\"^0.12.21\",\"karma-jasmine\":\"^0.1.5\",\"grunt-karma\":\"^0.8.3\",\"karma-phantomjs-launcher\":\"^0.1.4\",\"karma-jasmine-ajax\":\"^0.1.4\",\"grunt-update-json\":\"^0.1.3\",\"grunt-contrib-nodeunit\":\"^0.4.1\",\"grunt-banner\":\"^0.2.3\"},\"_id\":\"axios@0.2.0\",\"dist\":{\"shasum\":\"315cd618142078fd22f2cea35380caad19e32069\",\"tarball\":\"https://registry.npmjs.org/axios/-/axios-0.2.0.tgz\",\"integrity\":\"sha512-ZQb2IDQfop5Asx8PlKvccsSVPD8yFCwYZpXrJCyU+MqL4XgJVjMHkCTNQV/pmB0Wv7l74LUJizSM/SiPz6r9uw==\",\"signatures\":[{\"keyid\":\"SHA256:jl3bwswu80PjjokCgh0o2w5c2U4LhQAE57gj9cz1kzA\",\"sig\":\"MEQCIAkrijLTtL7uiw0fQf5GL/y7bJ+3J8Z0zrrzNLC5fTXlAiBd4Nr/EJ2nWfBGWv/9OkrAONoboG5C8t8plIt5LVeGQA==\"}]},\"_from\":\"./\",\"_npmVersion\":\"1.4.3\",\"_npmUser\":{\"name\":\"mzabriskie\",\"email\":\"mzabriskie@gmail.com\"},\"maintainers\":[{\"name\":\"mzabriskie\",\"email\":\"mzabriskie@gmail.com\"}],\"directories\":{},\"deprecated\":\"Critical security vulnerability fixed in v0.21.1. For more information, see https://github.com/axios/axios/pull/3410\"}},\"readme\":\"axios\",\"maintainers\":[],\"time\":{\"modified\":\"2022-12-29T06:38:42.456Z\",\"created\":\"2014-08-29T23:08:36.810Z\",\"0.1.0\":\"2014-08-29T23:08:36.810Z\",\"0.2.0\":\"2014-09-12T20:06:33.167Z\"},\"homepage\":\"https://axios-http.com\",\"keywords\":[],\"repository\":{\"type\":\"git\",\"url\":\"git+https://github.com/axios/axios.git\"},\"author\":{\"name\":\"Matt Zabriskie\"},\"bugs\":{\"url\":\"https://github.com/axios/axios/issues\"},\"license\":\"MIT\",\"readmeFilename\":\"README.md\",\"users\":{},\"contributors\":[]}\n"
const existingScopedPackageData = "{\"_id\":\"@TestScope/test\",\"_rev\":\"779-b37ceeb27a03858a89a0226f7c554aaf\",\"name\":\"@TestScope/test\",\"description\":\"Promise based HTTP client for the browser and node.js\",\"dist-tags\":{\"latest\":\"0.1.0\",\"next\":\"0.2.0\"},\"versions\":{\"0.1.0\":{\"name\":\"@TestScope/test\",\"version\":\"0.1.0\",\"description\":\"Promise based XHR library\",\"main\":\"index.js\",\"scripts\":{\"test\":\"grunt test\",\"start\":\"node ./sandbox/index.js\"},\"repository\":{\"type\":\"git\",\"url\":\"https://github.com/mzabriskie/@TestScope/test.git\"},\"keywords\":[\"xhr\",\"http\",\"ajax\",\"promise\"],\"author\":{\"name\":\"Matt Zabriskie\"},\"license\":\"MIT\",\"bugs\":{\"url\":\"https://github.com/mzabriskie/@TestScope/test/issues\"},\"homepage\":\"https://github.com/mzabriskie/@TestScope/test\",\"dependencies\":{\"es6-promise\":\"^1.0.0\"},\"devDependencies\":{\"grunt\":\"^0.4.5\",\"grunt-contrib-clean\":\"^0.6.0\",\"grunt-contrib-watch\":\"^0.6.1\",\"webpack\":\"^1.3.3-beta2\",\"webpack-dev-server\":\"^1.4.10\",\"grunt-webpack\":\"^1.0.8\",\"load-grunt-tasks\":\"^0.6.0\",\"karma\":\"^0.12.21\",\"karma-jasmine\":\"^0.1.5\",\"grunt-karma\":\"^0.8.3\",\"karma-phantomjs-launcher\":\"^0.1.4\",\"karma-jasmine-ajax\":\"^0.1.4\",\"grunt-update-json\":\"^0.1.3\",\"grunt-contrib-nodeunit\":\"^0.4.1\",\"grunt-banner\":\"^0.2.3\"},\"_id\":\"@TestScope/test@0.1.0\",\"dist\":{\"shasum\":\"854e14f2999c2ef7fab058654fd995dd183688f2\",\"tarball\":\"https://registry.npmjs.org/@TestScope/test/-/@TestScope/test-0.1.0.tgz\",\"integrity\":\"sha512-hRPotWTy88LEsJ31RWEs2fmU7mV2YJs3Cw7Tk5XkKGtnT5NKOyIvPU+6qTWfwQFusxzChe8ozjay8r56wfpX8w==\",\"signatures\":[{\"keyid\":\"SHA256:jl3bwswu80PjjokCgh0o2w5c2U4LhQAE57gj9cz1kzA\",\"sig\":\"MEYCIQC/cOvHsV7UqLAet6WE89O4Ga3AUHgkqqoP0riLs6sgTAIhAIrePavu3Uw0T3vLyYMlfEI9bqENYjPzH5jGK8vYQVJK\"}]},\"_from\":\"./\",\"_npmVersion\":\"1.4.3\",\"_npmUser\":{\"name\":\"mzabriskie\",\"email\":\"mzabriskie@gmail.com\"},\"maintainers\":[{\"name\":\"mzabriskie\",\"email\":\"mzabriskie@gmail.com\"}],\"directories\":{},\"deprecated\":\"Critical security vulnerability fixed in v0.21.1. For more information, see https://github.com/@TestScope/test/@TestScope/test/pull/3410\"},\"0.2.0\":{\"name\":\"@TestScope/test\",\"version\":\"0.2.0\",\"description\":\"Promise based HTTP client for the browser and node.js\",\"main\":\"index.js\",\"scripts\":{\"test\":\"grunt test\",\"start\":\"node ./sandbox/server.js\"},\"repository\":{\"type\":\"git\",\"url\":\"https://github.com/mzabriskie/@TestScope/test.git\"},\"keywords\":[\"xhr\",\"http\",\"ajax\",\"promise\",\"node\"],\"author\":{\"name\":\"Matt Zabriskie\"},\"license\":\"MIT\",\"bugs\":{\"url\":\"https://github.com/mzabriskie/@TestScope/test/issues\"},\"homepage\":\"https://github.com/mzabriskie/@TestScope/test\",\"dependencies\":{\"es6-promise\":\"^1.0.0\"},\"devDependencies\":{\"grunt\":\"^0.4.5\",\"grunt-contrib-clean\":\"^0.6.0\",\"grunt-contrib-watch\":\"^0.6.1\",\"webpack\":\"^1.3.3-beta2\",\"webpack-dev-server\":\"^1.4.10\",\"grunt-webpack\":\"^1.0.8\",\"load-grunt-tasks\":\"^0.6.0\",\"karma\":\"^0.12.21\",\"karma-jasmine\":\"^0.1.5\",\"grunt-karma\":\"^0.8.3\",\"karma-phantomjs-launcher\":\"^0.1.4\",\"karma-jasmine-ajax\":\"^0.1.4\",\"grunt-update-json\":\"^0.1.3\",\"grunt-contrib-nodeunit\":\"^0.4.1\",\"grunt-banner\":\"^0.2.3\"},\"_id\":\"@TestScope/test@0.2.0\",\"dist\":{\"shasum\":\"315cd618142078fd22f2cea35380caad19e32069\",\"tarball\":\"https://registry.npmjs.org/@TestScope/test/-/@TestScope/test-0.2.0.tgz\",\"integrity\":\"sha512-ZQb2IDQfop5Asx8PlKvccsSVPD8yFCwYZpXrJCyU+MqL4XgJVjMHkCTNQV/pmB0Wv7l74LUJizSM/SiPz6r9uw==\",\"signatures\":[{\"keyid\":\"SHA256:jl3bwswu80PjjokCgh0o2w5c2U4LhQAE57gj9cz1kzA\",\"sig\":\"MEQCIAkrijLTtL7uiw0fQf5GL/y7bJ+3J8Z0zrrzNLC5fTXlAiBd4Nr/EJ2nWfBGWv/9OkrAONoboG5C8t8plIt5LVeGQA==\"}]},\"_from\":\"./\",\"_npmVersion\":\"1.4.3\",\"_npmUser\":{\"name\":\"mzabriskie\",\"email\":\"mzabriskie@gmail.com\"},\"maintainers\":[{\"name\":\"mzabriskie\",\"email\":\"mzabriskie@gmail.com\"}],\"directories\":{},\"deprecated\":\"Critical security vulnerability fixed in v0.21.1. For more information, see https://github.com/@TestScope/test/@TestScope/test/pull/3410\"}},\"readme\":\"@TestScope/test\",\"maintainers\":[],\"time\":{\"modified\":\"2022-12-29T06:38:42.456Z\",\"created\":\"2014-08-29T23:08:36.810Z\",\"0.1.0\":\"2014-08-29T23:08:36.810Z\",\"0.2.0\":\"2014-09-12T20:06:33.167Z\"},\"homepage\":\"https://@TestScope/test-http.com\",\"keywords\":[],\"repository\":{\"type\":\"git\",\"url\":\"git+https://github.com/@TestScope/test/@TestScope/test.git\"},\"author\":{\"name\":\"Matt Zabriskie\"},\"bugs\":{\"url\":\"https://github.com/@TestScope/test/@TestScope/test/issues\"},\"license\":\"MIT\",\"readmeFilename\":\"README.md\",\"users\":{},\"contributors\":[]}\n"
const nonExistingPackageData = "{\"error\":\"Not found\"}"

func GetMockClient(baseUrl string, mockedToken string, mockedBody string, mockedHTTPStatusCode int) *httpclient.MockClient {
	return &httpclient.MockClient{
		DoFunc: func(req *http.Request) (*http.Response, error) {
			var statusCode int
			var httpError error
			var body string
			if !strings.HasPrefix(req.URL.String(), baseUrl) {
				statusCode = 404
				httpError = errors.New("not found")
			} else if req.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", mockedToken) {
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

func CreateDummyRc() (string, error) {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}
	config, err := os.Create(filepath.Join(dir, ".npmrc"))
	if err != nil {
		return "", err
	}
	defer config.Close()
	_, err = fmt.Fprintf(config, "//mycustomregistry.updatecli.io/:_authToken=mytoken\n")
	if err != nil {
		return "", err
	}
	_, err = fmt.Fprintf(config, "@TestScope:registry=https://mycustomregistry.updatecli.io/\n")
	if err != nil {
		return "", err
	}
	return dir, nil
}
