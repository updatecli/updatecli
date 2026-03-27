package updateclihttp

import (
	"bytes"
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
		want                  string
		wantStatus            string
		wantErr               error
		specErr               error
	}{
		{
			name: "Normal case with default index",
			spec: Spec{
				Url: "https://repo.jenkins-ci.org/releases/org/jenkins-ci/main/jenkins-war/maven-metadata.xml",
			},
			mockedHTTPStatusCode: http.StatusOK,
			mockedHTTPBody:       multiLineText,
			want:                 multiLineText,
			wantStatus:           result.SUCCESS,
		},
		{
			name: "Normal case with verb",
			spec: Spec{
				Url: "https://repo.jenkins-ci.org/releases/org/jenkins-ci/main/jenkins-war/maven-metadata.xml",
				Request: Request{
					Verb: "HEAD",
				},
			},
			mockedHTTPStatusCode: http.StatusOK,
			mockedHTTPBody:       multiLineText,
			want:                 multiLineText,
			wantStatus:           result.SUCCESS,
		},
		{
			name: "POST case with no body",
			spec: Spec{
				Url: "https://repo.jenkins-ci.org/releases/org/jenkins-ci/main/jenkins-war/maven-metadata.xml",
				Request: Request{
					Verb: "POST",
				},
			},
			specErr: fmt.Errorf("requires spec.body when using POST, PUT or PATCH method"),
		},
		{
			name: "POST case with body",
			spec: Spec{
				Url: "https://repo.jenkins-ci.org/releases/org/jenkins-ci/main/jenkins-war/maven-metadata.xml",
				Request: Request{
					Verb: "POST",
					Body: "input",
				},
			},
			mockedHTTPStatusCode: http.StatusOK,
			mockedHTTPBody:       multiLineText,
			want:                 multiLineText,
			wantStatus:           result.SUCCESS,
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
			want:       monoLineText,
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
		{
			name: "Normal case with single request header",
			spec: Spec{
				Url: "https://repo.jenkins-ci.org/releases/org/jenkins-ci/main/jenkins-war/maven-metadata.xml",
				Request: Request{
					Headers: map[string]string{
						"Authorization": "Bearer test-token",
					},
				},
			},
			mockedHTTPStatusCode: http.StatusOK,
			mockedHTTPBody:       multiLineText,
			want:                 multiLineText,
			wantStatus:           result.SUCCESS,
		},
		{
			name: "Normal case with multiple request headers",
			spec: Spec{
				Url: "https://repo.jenkins-ci.org/releases/org/jenkins-ci/main/jenkins-war/maven-metadata.xml",
				Request: Request{
					Headers: map[string]string{
						"Authorization": "Bearer test-token",
						"X-Custom-Header": "custom-value",
					},
				},
			},
			mockedHTTPStatusCode: http.StatusOK,
			mockedHTTPBody:       multiLineText,
			want:                 multiLineText,
			wantStatus:           result.SUCCESS,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sut, sutErr := New(tt.spec)
			if tt.specErr != nil {
				require.Error(t, sutErr)
				assert.Equal(t, tt.specErr, sutErr)
				return
			} else {
				require.NoError(t, sutErr)
			}

			sut.httpClient = &httpclient.MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					// check method
					if tt.spec.Request.Verb != "" &&
						tt.spec.Request.Verb != req.Method {
						return nil, fmt.Errorf("unexpected verb, expected %s, got %s", tt.spec.Request.Verb, req.Method)
					}
					// check headers
					for k, v := range tt.spec.Request.Headers {
						got := req.Header.Get(k)
						if got != v {
							return nil, fmt.Errorf("unexpected header %q value, expected %q, got %q", k, v, got)
						}
					}
					// check body
					if tt.spec.Request.Body != "" {
						if req.Body == nil {
							return nil, fmt.Errorf("missing request body")
						}
						buf := new(bytes.Buffer)
						_, err := buf.ReadFrom(req.Body)
						if err != nil {
							return nil, err
						}
						s := buf.String()
						if tt.spec.Request.Body != s {
							return nil, fmt.Errorf("unexpected body, expected %s, got %s", tt.spec.Request.Body, s)
						}
					}

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
