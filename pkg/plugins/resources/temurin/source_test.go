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
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestSource(t *testing.T) {
	tests := []struct {
		name                 string
		spec                 Spec
		workingDir           string
		mockedHTTPStatusCode map[string]int
		mockedReleases       []string
		mockedLocationHeader string
		mockedParsedVersion  parsedVersion
		mockedHttpError      error
		want                 string
		wantStatus           string
		wantErr              string
	}{
		{
			name:           "Normal case with latest LTS and default",
			mockedReleases: []string{"jdk-21.0.4+7", "jdk-21.0.3+8", "jdk-21.0.2+4"},
			spec:           Spec{},
			want:           "jdk-21.0.4+7",
			wantStatus:     result.SUCCESS,
		},
		{
			name:                 "Normal case with installer_url and defaults",
			mockedReleases:       []string{"jdk-21.0.4+7", "jdk-21.0.3+8", "jdk-21.0.2+4"},
			mockedLocationHeader: "https://github.com/adoptium/temurin21-binaries/releases/download/jdk-21.0.4%2B7/OpenJDK21U-jdk_x64_linux_hotspot_21.0.4_7.tar.gz",
			spec: Spec{
				Result: "installer_url",
			},
			want:       "https://github.com/adoptium/temurin21-binaries/releases/download/jdk-21.0.4%2B7/OpenJDK21U-jdk_x64_linux_hotspot_21.0.4_7.tar.gz",
			wantStatus: result.SUCCESS,
		},
		{
			name:                 "Normal case with installer_url and defaults",
			mockedReleases:       []string{"jdk-21.0.4+7", "jdk-21.0.3+8", "jdk-21.0.2+4"},
			mockedLocationHeader: "https://github.com/adoptium/temurin21-binaries/releases/download/jdk-21.0.4%2B7/OpenJDK21U-jdk_x64_linux_hotspot_21.0.4_7.tar.gz.sha256.txt",
			spec: Spec{
				Result: "checksum_url",
			},
			want:       "https://github.com/adoptium/temurin21-binaries/releases/download/jdk-21.0.4%2B7/OpenJDK21U-jdk_x64_linux_hotspot_21.0.4_7.tar.gz.sha256.txt",
			wantStatus: result.SUCCESS,
		},
		{
			name:                 "Normal case with installer_url and defaults",
			mockedReleases:       []string{"jdk-21.0.4+7", "jdk-21.0.3+8", "jdk-21.0.2+4"},
			mockedLocationHeader: "https://github.com/adoptium/temurin21-binaries/releases/download/jdk-21.0.4%2B7/OpenJDK21U-jdk_x64_linux_hotspot_21.0.4_7.tar.gz.sig",
			spec: Spec{
				Result: "signature_url",
			},
			want:       "https://github.com/adoptium/temurin21-binaries/releases/download/jdk-21.0.4%2B7/OpenJDK21U-jdk_x64_linux_hotspot_21.0.4_7.tar.gz.sig",
			wantStatus: result.SUCCESS,
		},
		{
			name:           "Normal case with latest LTS nightly (EA) release",
			mockedReleases: []string{"jdk-21.0.5+4-ea-beta", "jdk-21.0.4+6-ea-beta", "jdk-21.0.3+8-ea-beta"},
			spec: Spec{
				ReleaseLine: "feature",
				ReleaseType: "ea",
			},
			want:       "jdk-21.0.5+4-ea-beta",
			wantStatus: result.SUCCESS,
		},
		{
			name:           "Normal case with user specified version",
			mockedReleases: []string{"jdk-17.0.2+9"},
			mockedParsedVersion: parsedVersion{
				Major:    17,
				Minor:    0,
				Security: 2,
			},
			spec: Spec{
				SpecificVersion: "17.0.2",
			},
			want:       "jdk-17.0.2+9",
			wantStatus: result.SUCCESS,
		},
		{
			name:            "Failure when HTTP request error",
			mockedHttpError: fmt.Errorf("Connection Error"),
			spec:            Spec{},
			wantStatus:      result.FAILURE,
			wantErr:         "something went wrong while performing a request to \"https://api.adoptium.net/v3/info/available_releases\":\nConnection Error\n",
		},
		{
			name:                 "Failure when HTTP/500 from the 'available_releases' endpoint",
			mockedHTTPStatusCode: map[string]int{availableReleasesEndpoint: http.StatusInternalServerError},
			spec:                 Spec{},
			wantStatus:           result.FAILURE,
			wantErr:              fmt.Sprintf("Got an HTTP error %d from the API.\n", http.StatusInternalServerError),
		},
		{
			name:                 "Failure when HTTP/500 from the 'release_names' endpoint",
			mockedHTTPStatusCode: map[string]int{releaseNamesEndpoint: http.StatusInternalServerError},
			spec:                 Spec{},
			wantStatus:           result.FAILURE,
			wantErr:              fmt.Sprintf("Got an HTTP error %d from the API.\n", http.StatusInternalServerError),
		},
		{
			name:                 "Failure when HTTP/500 from the '" + installersEndpoint + "' endpoint",
			mockedReleases:       []string{"jdk-21.0.5+4-ea-beta", "jdk-21.0.4+6-ea-beta", "jdk-21.0.3+8-ea-beta"},
			mockedHTTPStatusCode: map[string]int{installersEndpoint: http.StatusInternalServerError},
			spec: Spec{
				Result: "installer_url",
			},
			wantStatus: result.FAILURE,
			wantErr:    fmt.Sprintf("Got an HTTP error %d from the API.\n", http.StatusInternalServerError),
		},
		{
			name: "Failure when HTTP/500 from the '/version' endpoint",
			// mockedReleases:        []string{"jdk-21.0.5+4-ea-beta", "jdk-21.0.4+6-ea-beta", "jdk-21.0.3+8-ea-beta"},
			mockedHTTPStatusCode: map[string]int{parseVersionEndpoint: http.StatusInternalServerError},
			spec: Spec{
				SpecificVersion: "17.0.2",
			},
			wantStatus: result.FAILURE,
			wantErr:    "the version \"17.0.2\" is not a valid Temurin version.\nAPI response was: \"Got an HTTP error 500 from the API.\\n\"",
		},
		{
			name:                 "Failure when HTTP/500 from the 'checksum/version' endpoint",
			mockedReleases:       []string{"jdk-21.0.5+4-ea-beta", "jdk-21.0.4+6-ea-beta", "jdk-21.0.3+8-ea-beta"},
			mockedHTTPStatusCode: map[string]int{checksumsEndpoint: http.StatusInternalServerError},
			spec: Spec{
				Result: "checksum_url",
			},
			wantStatus: result.FAILURE,
			wantErr:    fmt.Sprintf("Got an HTTP error %d from the API.\n", http.StatusInternalServerError),
		},
		{
			name:                 "Failure when HTTP/500 from the 'signature/version' endpoint",
			mockedReleases:       []string{"jdk-21.0.5+4-ea-beta", "jdk-21.0.4+6-ea-beta", "jdk-21.0.3+8-ea-beta"},
			mockedHTTPStatusCode: map[string]int{signaturesEndpoint: http.StatusInternalServerError},
			spec: Spec{
				Result: "signature_url",
			},
			wantStatus: result.FAILURE,
			wantErr:    fmt.Sprintf("Got an HTTP error %d from the API.\n", http.StatusInternalServerError),
		},
		{
			name:           "Failure when 'release_names' provides empty results",
			mockedReleases: []string{},
			spec:           Spec{},
			wantStatus:     result.FAILURE,
			wantErr:        "[temurin] No release found matching provided criteria. Use '--debug' to get details.",
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

			got := result.Source{}
			gotErr := sut.Source(tt.workingDir, &got)

			if tt.wantErr != "" {
				require.Error(t, gotErr)
				// assert.Equal(t, got.Result, result.FAILURE)
				assert.Equal(t, tt.wantErr, gotErr.Error())
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantStatus, got.Result)
			assert.Equal(t, tt.want, got.Information)
		})
	}
}
