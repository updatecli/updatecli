package registry

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
)

func TestCondition(t *testing.T) {
	tests := []struct {
		name             string
		spec             Spec
		source           string
		expectedResult   bool
		expectedError    bool
		expectedErrorMsg error
		expectedUrl      string
		mockedHttpBody   string
		mockedHttpError  error
	}{
		{
			name: "Success - provider",
			spec: Spec{
				Type:      "provider",
				Namespace: "hashicorp",
				Name:      "kubernetes",
				Version:   "2.23.0",
			},
			expectedResult: true,
			expectedUrl:    "https://registry.terraform.io/v1/providers/hashicorp/kubernetes",
			mockedHttpBody: `{ "versions" : ["2.23.0"] }`,
		},
		{
			name: "Success - module",
			spec: Spec{
				Type:         "module",
				Namespace:    "terraform-aws-modules",
				Name:         "vpc",
				TargetSystem: "aws",
				Version:      "5.1.1",
			},
			expectedResult: true,
			expectedUrl:    "https://registry.terraform.io/v1/modules/terraform-aws-modules/vpc/aws",
			mockedHttpBody: `{ "versions" : ["5.1.1"] }`,
		},
		{
			name: "Success - provider source",
			spec: Spec{
				Type:      "provider",
				Namespace: "hashicorp",
				Name:      "kubernetes",
			},
			source:         "2.23.0",
			expectedResult: true,
			expectedUrl:    "https://registry.terraform.io/v1/providers/hashicorp/kubernetes",
			mockedHttpBody: `{ "versions" : ["2.23.0"] }`,
		},
		{
			name: "Success - module source",
			spec: Spec{
				Type:         "module",
				Namespace:    "terraform-aws-modules",
				Name:         "vpc",
				TargetSystem: "aws",
			},
			source:         "5.1.1",
			expectedResult: true,
			expectedUrl:    "https://registry.terraform.io/v1/modules/terraform-aws-modules/vpc/aws",
			mockedHttpBody: `{ "versions" : ["5.1.1"] }`,
		},
		{
			name: "Failed - missing version",
			spec: Spec{
				Type:      "provider",
				Namespace: "hashicorp",
				Name:      "kubernetes",
				Version:   "2.22.1111",
			},
			expectedResult: false,
			expectedUrl:    "https://registry.terraform.io/v1/providers/hashicorp/kubernetes",
			mockedHttpBody: `{ "versions" : ["2.23.0"] }`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			require.NoError(t, err)

			got, err := New(tt.spec)
			require.NoError(t, err)

			got.webClient = &httpclient.MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, tt.expectedUrl, req.URL.String())
					body := tt.mockedHttpBody
					statusCode := 200
					return &http.Response{
						StatusCode: statusCode,
						Body:       io.NopCloser(strings.NewReader(body)),
					}, tt.mockedHttpError
				},
			}

			gotResult, _, gotErr := got.Condition(tt.source, nil)
			if tt.expectedError {
				if assert.Error(t, gotErr) {
					assert.Equal(t, tt.expectedErrorMsg.Error(), gotErr.Error())
				}
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
