package updateclihttp

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestCondition(t *testing.T) {
	tests := []struct {
		name                  string
		spec                  Spec
		source                string
		scm                   scm.ScmHandler
		mockedHTTPStatusCode  int
		mockedHTTPBody        string
		mockedHTTPRespHeaders http.Header
		mockedHttpError       error
		want                  bool
		wantErr               error
	}{
		{
			name: "Success case with existing URL",
			spec: Spec{
				Url: "https://repo.jenkins-ci.org/releases/org/jenkins-ci/main/jenkins-war/maven-metadata.xml",
			},
			mockedHTTPStatusCode: http.StatusOK,
			want:                 true,
		},
		{
			name: "Success case with custom request and existing URL",
			spec: Spec{
				Url: "https://repo.jenkins-ci.org/releases/org/jenkins-ci/main/jenkins-war/maven-metadata.xml",
				Request: Request{
					Verb: "HEAD",
					Headers: map[string]string{
						"Authorization": "Bearer Token",
					},
				},
			},
			mockedHTTPStatusCode: http.StatusOK,
			want:                 true,
		},
		{
			name: "Success case with assertions",
			spec: Spec{
				Url: "https://repo.jenkins-ci.org/releases/org/jenkins-ci/main/jenkins-war/nope",
				ResponseAsserts: ResponseAsserts{
					StatusCode: 500,
					Headers: map[string]string{
						"Location": "https://google.com",
					},
				},
			},
			mockedHTTPStatusCode: http.StatusInternalServerError,
			mockedHTTPRespHeaders: http.Header{
				"Location":     {"https://google.com"},
				"Content-Type": {"application/xml"},
			},
			want: true,
		},
		{
			name: "Failing case with not-existing URL",
			spec: Spec{
				Url: "https://repo.jenkins-ci.org/releases/org/jenkins-ci/main/jenkins-war/nope",
			},
			mockedHTTPStatusCode: http.StatusNotFound,
			want:                 false,
		},
		{
			name: "Failing case with unmet assertion on status code",
			spec: Spec{
				Url: "https://repo.jenkins-ci.org/releases/org/jenkins-ci/main/jenkins-war/maven-metadata.xml",
				ResponseAsserts: ResponseAsserts{
					StatusCode: 500,
					Headers: map[string]string{
						"Location": "https://google.com",
					},
				},
			},
			mockedHTTPStatusCode: http.StatusOK,
			mockedHTTPRespHeaders: http.Header{
				"Location":     {"https://google.com"},
				"Content-Type": {"application/xml"},
			},
			want: false,
		},
		{
			name: "Failing case with unmet assertion on headers code",
			spec: Spec{
				Url: "https://repo.jenkins-ci.org/releases/org/jenkins-ci/main/jenkins-war/maven-metadata.xml",
				ResponseAsserts: ResponseAsserts{
					StatusCode: 200,
					Headers: map[string]string{
						"Location": "https://google.com",
					},
				},
			},
			mockedHTTPStatusCode: http.StatusOK,
			want:                 false,
		},
		{
			name: "Error (and failing) case when HTTP code is >= 500",
			spec: Spec{
				Url: "https://repo.jenkins-ci.org/releases/org/jenkins-ci/main/jenkins-war/maven-metadata.xml",
			},
			mockedHTTPStatusCode: http.StatusInternalServerError,
			wantErr:              &ErrHttpError{resStatusCode: http.StatusInternalServerError},
			want:                 false,
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

			got := &result.Condition{}
			gotErr := sut.Condition(tt.source, tt.scm, got)

			if tt.wantErr != nil {
				require.Error(t, gotErr)
				assert.Equal(t, got.Result, result.FAILURE)
				assert.Equal(t, tt.wantErr, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.want, got.Pass)

			if tt.want {
				assert.Equal(t, got.Result, result.SUCCESS)
			} else {
				assert.Equal(t, got.Result, result.FAILURE)
			}
		})
	}
}
