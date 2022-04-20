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

func Test_New(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		username string
		password string
		want     Registry
	}{
		{
			name:     "GHCR.io with a bearer token",
			hostname: "ghcr.io",
			username: "joe",
			password: "TopSecret2020!",
			want: DockerGenericRegistry{
				Auth: RegistryAuth{
					Username: "joe",
					Password: "TopSecret2020!",
				},
				WebClient: http.DefaultClient,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.hostname, tt.username, tt.password)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRegistry_Digest(t *testing.T) {
	tests := []struct {
		name     string
		image    dockerimage.Image
		username string
		password string
		want     string
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
			name: "Normal case with anonymous request on GHCR",
			image: dockerimage.Image{
				Registry:     "ghcr.io",
				Namespace:    "joe",
				Repository:   "updatecli",
				Tag:          "latest",
				Architecture: "arm64",
			},
			want: "c74f1b1166784193ea6c8f9440263b9be6cae07dfe35e32a5df7a31358ac2060",
		},
		{
			name: "Normal case with authenticated request on DockerHub",
			image: dockerimage.Image{
				Registry:     "hub.docker.com",
				Namespace:    "joe",
				Repository:   "updatecli",
				Tag:          "latest",
				Architecture: "arm64",
			},
			username: "ava",
			password: "supersecretpassword",
			want:     "c74f1b1166784193ea6c8f9440263b9be6cae07dfe35e32a5df7a31358ac2060",
		},
		{
			name: "Normal case on quay.io (new OCI format: manifest list)",
			image: dockerimage.Image{
				Registry:     "quay.io",
				Namespace:    "ansible",
				Repository:   "ansible-runner",
				Tag:          "latest",
				Architecture: "arm64",
			},
			mockAPIResHeaders: http.Header{
				"Content-Type": {"application/vnd.oci.image.index.v1+json"},
			},
			want: "c74f1b1166784193ea6c8f9440263b9be6cae07dfe35e32a5df7a31358ac2060",
		},
		{
			name: "Normal case on quay.io (new OCI format: standalone manifest)",
			image: dockerimage.Image{
				Registry:     "quay.io",
				Namespace:    "ansible",
				Repository:   "ansible-runner",
				Tag:          "latest",
				Architecture: "arm64",
			},
			mockAPIResHeaders: http.Header{
				"Content-Type":          {"application/vnd.oci.image.manifest.v1+json"},
				"Docker-Content-Digest": {"sha256:abb5ef7d2825f8ca4927f406cce339ca3b66d29f6267c26234255546680642c3"},
			},
			want: "abb5ef7d2825f8ca4927f406cce339ca3b66d29f6267c26234255546680642c3",
		},
		{
			name: "Image does not exist",
			image: dockerimage.Image{
				Registry:     "ghcr.io",
				Namespace:    "joe",
				Repository:   "updatecli",
				Tag:          "donotexist",
				Architecture: "arm64",
			},
			mockAPIStatusCode: 404,
			want:              "",
			wantErr:           false,
		},
		{
			name: "Unauthenticated on Registry",
			image: dockerimage.Image{
				Registry:     "ghcr.io",
				Namespace:    "joe",
				Repository:   "nginx",
				Tag:          "updatecli",
				Architecture: "arm64",
			},
			mockAPIStatusCode: 401,
			wantErr:           true,
		},
		{
			name: "HTTP general error",
			image: dockerimage.Image{
				Registry:     "ghcr.io",
				Namespace:    "joe",
				Repository:   "nginx",
				Tag:          "updatecli",
				Architecture: "arm64",
			},
			mockAPIError: fmt.Errorf("dial tcp hub.docker.com: i/o timeout"),
			want:         "",
			wantErr:      true,
		},
		{
			name: "Cannot parse JSON from response body",
			image: dockerimage.Image{
				Registry:   "ghcr.io",
				Namespace:  "joe",
				Repository: "nginx",
				Tag:        "updatecli",
			},
			mockAPIResBody: `{[,`,
			want:           "",
			wantErr:        true,
		},
		{
			name: "Unsupported API content type",
			image: dockerimage.Image{
				Registry:   "localhost:5000",
				Namespace:  "joe",
				Repository: "ansible",
				Tag:        "latest",
			},
			mockAPIResHeaders: http.Header{
				"Content-Type": {"application/vnd.docker.distribution.manifest.v1+prettyjws"},
			},
			want:    "",
			wantErr: true,
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
							apiBody = exampleManifest
						}

						statusCode := tt.mockAPIStatusCode
						if statusCode == 0 {
							statusCode = 200
						}

						headers := tt.mockAPIResHeaders
						if headers == nil || len(headers) == 0 {
							headers = http.Header{"Content-Type": {"application/vnd.docker.distribution.manifest.list.v2+json"}}
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

			got, gotErr := sut.Digest(tt.image)
			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Content from Dockerhub at https://hub.docker.com/v2/repositories/library/nginx/tags/latest
const exampleManifest = `{
  "schemaVersion": 2,
  "mediaType": "application/vnd.docker.distribution.manifest.list.v2+json",
  "manifests": [
    {
      "digest": "sha256:e7d88de73db3d3fd9b2d63aa7f447a10fd0220b7cbf39803c803f2af9ba256b3",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "amd64",
        "os": "linux"
      },
      "size": 528
    },
    {
      "digest": "sha256:e047bc2af17934d38c5a7fa9f46d443f1de3a7675546402592ef805cfa929f9d",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "arm",
        "os": "linux",
        "variant": "v6"
      },
      "size": 528
    },
    {
      "digest": "sha256:8483ecd016885d8dba70426fda133c30466f661bb041490d525658f1aac73822",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "arm",
        "os": "linux",
        "variant": "v7"
      },
      "size": 528
    },
    {
      "digest": "sha256:c74f1b1166784193ea6c8f9440263b9be6cae07dfe35e32a5df7a31358ac2060",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "arm64",
        "os": "linux",
        "variant": "v8"
      },
      "size": 528
    },
    {
      "digest": "sha256:2689e157117d2da668ad4699549e55eba1ceb79cb7862368b30919f0488213f4",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "386",
        "os": "linux"
      },
      "size": 528
    },
    {
      "digest": "sha256:2042a492bcdd847a01cd7f119cd48caa180da696ed2aedd085001a78664407d6",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "ppc64le",
        "os": "linux"
      },
      "size": 528
    },
    {
      "digest": "sha256:49e322ab6690e73a4909f787bcbdb873631264ff4a108cddfd9f9c249ba1d58e",
      "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
      "platform": {
        "architecture": "s390x",
        "os": "linux"
      },
      "size": 528
    }
  ]
}`
