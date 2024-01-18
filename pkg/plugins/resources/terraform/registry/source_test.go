package registry

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

func TestSource(t *testing.T) {
	tests := []struct {
		name            string
		spec            Spec
		expectedResult  string
		expectedError   bool
		mockedHttpBody  string
		mockedHttpError error
	}{
		{
			name: "Success - provider",
			spec: Spec{
				Type:      "provider",
				Namespace: "hashicorp",
				Name:      "kubernetes",
			},
			expectedResult: "2.23.0",
			mockedHttpBody: `{ "versions" : ["2.23.0"] }`,
		},
		{
			name: "Success - provider",
			spec: Spec{
				Type:         "module",
				Namespace:    "terraform-aws-modules",
				Name:         "vpc",
				TargetSystem: "aws",
			},
			expectedResult: "5.1.1",
			mockedHttpBody: `{ "versions" : ["5.1.1"] }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			require.NoError(t, err)

			got.webClient = &httpclient.MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					body := tt.mockedHttpBody
					statusCode := 200
					return &http.Response{
						StatusCode: statusCode,
						Body:       io.NopCloser(strings.NewReader(body)),
					}, tt.mockedHttpError
				},
			}

			gotResult := result.Source{}
			err = got.Source("", &gotResult)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Information)
		})
	}
}
