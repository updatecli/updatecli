package registry

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/httpclient"
)

func TestRegistryAddressAPI(t *testing.T) {
	tests := []struct {
		name            string
		spec            Spec
		expectedResult  string
		expectedUrl     string
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
			expectedResult: "https://registry.terraform.io/v1/providers/hashicorp/kubernetes",
			expectedUrl:    "https://registry.terraform.io/.well-known/terraform.json",
			mockedHttpBody: `{"modules.v1":"/v1/modules/","providers.v1":"/v1/providers/"}`,
		},
		{
			name: "Success - provider raw",
			spec: Spec{
				Type:      "provider",
				RawString: "registry.terraform.io/hashicorp/kubernetes",
			},
			expectedResult: "https://registry.terraform.io/v1/providers/hashicorp/kubernetes",
			expectedUrl:    "https://registry.terraform.io/.well-known/terraform.json",
			mockedHttpBody: `{"modules.v1":"/v1/modules/","providers.v1":"/v1/providers/"}`,
		},
		{
			name: "Success - module",
			spec: Spec{
				Type:         "module",
				Namespace:    "terraform-aws-modules",
				Name:         "vpc",
				TargetSystem: "aws",
			},
			expectedResult: "https://registry.terraform.io/v1/modules/terraform-aws-modules/vpc/aws",
			expectedUrl:    "https://registry.terraform.io/.well-known/terraform.json",
			mockedHttpBody: `{"modules.v1":"/v1/modules/","providers.v1":"/v1/providers/"}`,
		},
		{
			name: "Success - module raw",
			spec: Spec{
				Type:      "module",
				RawString: "registry.terraform.io/terraform-aws-modules/vpc/aws",
			},
			expectedResult: "https://registry.terraform.io/v1/modules/terraform-aws-modules/vpc/aws",
			expectedUrl:    "https://registry.terraform.io/.well-known/terraform.json",
			mockedHttpBody: `{"modules.v1":"/v1/modules/","providers.v1":"/v1/providers/"}`,
		},
		{
			name: "Success - module with hostname",
			spec: Spec{
				Type:         "module",
				Hostname:     "gitlab.example.com",
				Namespace:    "updatecli",
				Name:         "namespace",
				TargetSystem: "kubernetes",
			},
			expectedResult: "https://gitlab.example.com/api/v4/packages/terraform/modules/v1/updatecli/namespace/kubernetes",
			expectedUrl:    "https://gitlab.example.com/.well-known/terraform.json",
			mockedHttpBody: `{"modules.v1":"/api/v4/packages/terraform/modules/v1/"}`,
		},
		{
			name: "Success - module raw with hostname",
			spec: Spec{
				Type:      "module",
				RawString: "gitlab.example.com/updatecli/namespace/kubernetes",
			},
			expectedResult: "https://gitlab.example.com/api/v4/packages/terraform/modules/v1/updatecli/namespace/kubernetes",
			expectedUrl:    "https://gitlab.example.com/.well-known/terraform.json",
			mockedHttpBody: `{"modules.v1":"/api/v4/packages/terraform/modules/v1/"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()

			require.NoError(t, err)

			webClient := &httpclient.MockClient{
				DoFunc: func(req *http.Request) (*http.Response, error) {
					assert.Equal(t, tt.expectedUrl, req.URL.String())
					body := tt.mockedHttpBody
					statusCode := 200
					return &http.Response{
						StatusCode: statusCode,
						Body:       ioutil.NopCloser(strings.NewReader(body)),
					}, tt.mockedHttpError
				},
			}

			got, err := newRegistryAddress(webClient, tt.spec)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedResult, got.API())
		})
	}
}
