package updateclihttp

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/result"
)

const (
	multiLineText string = `
<?xml version="1.0" encoding="UTF-8"?>
<metadata>
	<groupId>org.jenkins-ci.main</groupId>
	<artifactId>jenkins-war</artifactId>
	<versioning>
		<latest>2.432</latest>
		<release>2.426.1</release>
		<versions>
			<version>2.432</version>
		</versions>
		<lastUpdated>20231115143950</lastUpdated>
	</versioning>
</metadata>`
	monoLineText string = "https://azcopyvnext.azureedge.net/releases/release-10.21.2-20231106/azcopy_linux_amd64_10.21.2.tar.gz"
)

func TestSource(t *testing.T) {
	tests := []struct {
		name                  string
		spec                  Spec
		workingDir            string
		mockedHTTPStatusCode  int
		mockedHTTPBody        string
		mockedHTTPRespHeaders http.Header
		mockedHttpError       error
		want                  []result.SourceInformation
		wantStatus            string
		wantErr               error
	}{
		{
			name: "Normal case with default index",
			spec: Spec{
				Url: "https://repo.jenkins-ci.org/releases/org/jenkins-ci/main/jenkins-war/maven-metadata.xml",
			},
			mockedHTTPStatusCode: http.StatusOK,
			mockedHTTPBody:       multiLineText,
			want: []result.SourceInformation{{
				Value: multiLineText,
			}},
			wantStatus: result.SUCCESS,
		},
		{
			name: "Normal case with header as result",
			spec: Spec{
				Url:                  "https://repo.jenkins-ci.org/releases/org/jenkins-ci/main/jenkins-war/maven-metadata.xml",
				ReturnResponseHeader: "Location",
				Request: Request{
					Verb: "HEAD",
				},
			},
			mockedHTTPStatusCode: http.StatusPermanentRedirect,
			mockedHTTPRespHeaders: map[string][]string{
				"Location": {monoLineText},
			},
			want: []result.SourceInformation{{
				Value: monoLineText,
			}},
			wantStatus: result.SUCCESS,
		},
		{
			name: "Error when HTTP code is >= 400",
			spec: Spec{
				Url: "https://repo.jenkins-ci.org/releases/org/jenkins-ci/main/jenkins-war/maven-metadata.xml",
			},
			mockedHTTPStatusCode: http.StatusInternalServerError,
			wantErr:              &ErrHttpError{resStatusCode: http.StatusInternalServerError},
			wantStatus:           result.FAILURE,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut, sutErr := New(tt.spec)
			require.NoError(t, sutErr)

			sut.httpClient = &httpclient.MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					body := tt.mockedHTTPBody
					statusCode := tt.mockedHTTPStatusCode
					return &http.Response{
						StatusCode: statusCode,
						Body:       io.NopCloser(strings.NewReader(body)),
						Header:     tt.mockedHTTPRespHeaders,
					}, tt.mockedHttpError
				},
			}

			got := result.Source{}
			gotErr := sut.Source(tt.workingDir, &got)

			if tt.wantErr != nil {
				require.Error(t, gotErr)
				assert.Equal(t, got.Result, result.FAILURE)
				assert.Equal(t, tt.wantErr, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantStatus, got.Result)
			assert.Equal(t, tt.want, got.Information)
		})
	}
}
