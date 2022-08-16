package dockerregistry

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker/dockerimage"
)

const exampleTagListResponse = `{
	"name": "updatecli/updatecli",
	"tags": [ 
		"0.1.0",
		"0.2.0",
		"0.3.0"
	]
}`

func Test_Tags(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		username string
		password string
		image    dockerimage.Image
		want     []string
		wantErr  bool
		// provides a custom HTTP response code from the API service. Set to 200 if 0/unspecified.
		mockAPIStatusCode int
		// provides a custom HTTP response error from the API service. Set to "No error" if nil/unspecified.
		mockAPIError error
		// provides a custom HTTP response body from the API service. Set to the constant 'exampleManifest' if empty/unspecified.
		mockAPIResBody string
		// provides a custom HTTP response headers from the API service. Set to {"Content-Type": {"application/vnd.docker.distribution.manifest.list.v2+json"}} if nil/unspecified.
		mockAPIResHeaders http.Header
		// Custom mock HTTP response function for edge cases (is nil most of the time)
		mockFunc func(req *http.Request) (*http.Response, error)
	}{
		{
			name:     "GHCR.io with a bearer token",
			hostname: "ghcr.io",
			username: "john",
			password: "TopSecrets2020!",
			want: []string{
				"0.1.0",
				"0.2.0",
				"0.3.0",
			},
			image: dockerimage.Image{
				Registry:     "ghcr.io",
				Namespace:    "updatecli",
				Repository:   "updatecli",
				Architecture: "amd64",
			},
		},
		{
			name:     "GHCR.io with a bearer token",
			hostname: "ghcr.io",
			username: "john",
			password: "TopSecrets2020!",
			want:     []string{},
			image: dockerimage.Image{
				Registry:     "ghcr.io",
				Namespace:    "updatecli",
				Repository:   "updatecli",
				Architecture: "amd64",
			},
			mockAPIStatusCode: 404,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut := DockerGenericRegistry{
				// Mocking the HTTP client with this function to control the responses and introspect generated requests
				WebClient: &httpclient.MockClient{
					DoFunc: func(req *http.Request) (*http.Response, error) {
						// If a token service is called, then returns a mocked response with a token
						// The token is ALWAYS the following: "bearertoken-<provided username>-<provided password>"
						if strings.Contains(req.URL.String(), registryEndpoints(tt.image).TokenService) {
							// Returns a token with the username and password to allow introspection
							user, pass, _ := req.BasicAuth()
							body := fmt.Sprintf(`{"token":"bearertoken-%s-%s"}`, user, pass)
							return &http.Response{
								StatusCode: 200,
								Body:       ioutil.NopCloser(strings.NewReader(body)),
							}, nil
						}

						// Extract developer-provided values or set to defaults
						apiBody := tt.mockAPIResBody
						if apiBody == "" {
							apiBody = exampleTagListResponse
						}

						statusCode := tt.mockAPIStatusCode
						if statusCode == 0 {
							statusCode = 200
						}

						headers := tt.mockAPIResHeaders
						if len(headers) == 0 {
							headers = http.Header{"Content-Type": {"application/json"}}
						}

						// If the API service is called, returns the mocked answers
						return &http.Response{
							StatusCode: statusCode,
							Body:       ioutil.NopCloser(strings.NewReader(apiBody)),
							Header:     headers,
						}, tt.mockAPIError
					},
				},
				Auth: RegistryAuth{
					Username: tt.username,
					Password: tt.password,
				},
			}

			got, gotErr := sut.Tags(tt.image)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.want, got)

		})
	}
}
