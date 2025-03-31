package temurin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func TestCondition(t *testing.T) {
	tests := []struct {
		name                 string
		spec                 Spec
		source               string
		scm                  scm.ScmHandler
		mockedHTTPStatusCode map[string]int
		mockedReleases       []string
		mockedLocationHeader string
		mockedParsedVersion  parsedVersion
		mockedHttpError      error
		want                 bool
		wantErr              string
	}{
		{
			name:           "Normal case with latest LTS and defaults",
			mockedReleases: []string{"jdk-21.0.4+7", "jdk-21.0.3+8", "jdk-21.0.2+4"},
			spec:           Spec{},
			want:           true,
		},
		{
			name:           "Normal case with user specified version",
			mockedReleases: []string{"jdk-17.0.3+12", "jdk-17.0.2+9", "jdk-17.0.1+8"},
			mockedParsedVersion: parsedVersion{
				Major:    17,
				Minor:    0,
				Security: 2,
			},
			spec: Spec{
				SpecificVersion: "jdk-17.0.2+9",
			},
			want: true,
		},
		{
			name:           "Normal case with latest LTS, defaults and a source input",
			mockedReleases: []string{"jdk-21.0.4+7", "jdk-21.0.3+8", "jdk-21.0.2+4"},
			source:         "jdk-21.0.4+7",
			spec:           Spec{},
			want:           true,
		},
		{
			name:           "Normal case with latest LTS, list of platforms and a source input",
			mockedReleases: []string{"jdk-21.0.4+7", "jdk-21.0.3+8", "jdk-21.0.2+4"},
			source:         "jdk-21.0.4+7",
			spec: Spec{
				Platforms: []string{"linux/x64", "windows/x64"},
			},
			want: true,
		},
		{
			name:           "Failing case with no release found matching user provided Spec",
			mockedReleases: []string{},
			spec:           Spec{},
			want:           false,
		},
		{
			name:           "Failing case with user specified version not existing",
			mockedReleases: []string{"jdk-17.0.2+9", "jdk-21.0.4+7", "jdk-21.0.3+8", "jdk-21.0.2+4"},
			mockedParsedVersion: parsedVersion{
				Major:    17,
				Minor:    1,
				Security: 9,
			},
			spec: Spec{
				SpecificVersion: "17.1.9",
			},
			want: false,
		},
		{
			name:            "Failure when HTTP request error",
			mockedHttpError: fmt.Errorf("Connection Error"),
			spec:            Spec{},
			wantErr:         "something went wrong while performing a request to \"https://api.adoptium.net/v3/info/available_releases\":\nConnection Error",
		},
		{
			name:                 "Failure when HTTP/500 from the 'available_releases' endpoint",
			mockedHTTPStatusCode: map[string]int{availableReleasesEndpoint: http.StatusInternalServerError},
			spec:                 Spec{},
			wantErr:              fmt.Sprintf("got an HTTP error %d from the API", http.StatusInternalServerError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut, sutErr := New(tt.spec)
			require.NoError(t, sutErr)

			var mockedHttpClient = &httpclient.MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {

					mockedResponse := &http.Response{
						StatusCode: http.StatusOK,
					}
					mockedBody := []byte{}

					requestedEndpoint := req.URL.String()
					switch true {
					case strings.HasPrefix(requestedEndpoint, temurinApiUrl+availableReleasesEndpoint):
						mockedBody, _ = json.Marshal(apiInfoReleases{
							MostRecentLTS:            21,
							AvailableLTSReleases:     []int{21, 17, 11},
							MostRecentFeatureRelease: 23,
							AvailableReleases:        []int{23, 22, 21, 17, 11, 8},
						})
						mockedResponse.StatusCode = tt.mockedHTTPStatusCode[availableReleasesEndpoint]

					case strings.HasPrefix(requestedEndpoint, temurinApiUrl+releaseNamesEndpoint):
						mockedBody, _ = json.Marshal(releaseInformation{tt.mockedReleases})
						mockedResponse.StatusCode = tt.mockedHTTPStatusCode[releaseNamesEndpoint]

					case strings.HasPrefix(requestedEndpoint, temurinApiUrl+installersEndpoint):
						mockedResponse.Header = map[string][]string{
							"Location": {tt.mockedLocationHeader},
						}
						mockedResponse.StatusCode = tt.mockedHTTPStatusCode[installersEndpoint]

					case strings.HasPrefix(requestedEndpoint, temurinApiUrl+architecturesEndpoint):
						mockedResponse.StatusCode = tt.mockedHTTPStatusCode[architecturesEndpoint]

					case strings.HasPrefix(requestedEndpoint, temurinApiUrl+osEndpoints):
						mockedResponse.StatusCode = tt.mockedHTTPStatusCode[osEndpoints]

					case strings.HasPrefix(requestedEndpoint, temurinApiUrl+checksumsEndpoint):
						mockedResponse.Header = map[string][]string{
							"Location": {tt.mockedLocationHeader},
						}
						mockedResponse.StatusCode = tt.mockedHTTPStatusCode[checksumsEndpoint]

					case strings.HasPrefix(requestedEndpoint, temurinApiUrl+signaturesEndpoint):
						mockedResponse.Header = map[string][]string{
							"Location": {tt.mockedLocationHeader},
						}
						mockedResponse.StatusCode = tt.mockedHTTPStatusCode[signaturesEndpoint]

					case strings.HasPrefix(requestedEndpoint, temurinApiUrl+parseVersionEndpoint):
						mockedBody, _ = json.Marshal(tt.mockedParsedVersion)
						mockedResponse.StatusCode = tt.mockedHTTPStatusCode[parseVersionEndpoint]
					}

					mockedResponse.Body = io.NopCloser(bytes.NewReader(mockedBody))

					return mockedResponse, tt.mockedHttpError
				},
			}

			sut.apiWebClient = mockedHttpClient
			sut.apiWebRedirectionClient = mockedHttpClient

			gotResult, _, gotErr := sut.Condition(tt.source, tt.scm)

			if tt.wantErr != "" {
				require.Error(t, gotErr)
				assert.Equal(t, tt.wantErr, gotErr.Error())
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.want, gotResult)
		})
	}
}
