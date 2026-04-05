package pypi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestCondition(t *testing.T) {
	tests := []struct {
		name                 string
		spec                 Spec
		sourceInput          string
		expectedPass         bool
		expectedError        bool
		mockedBody           string
		mockedURL            string
		mockedToken          string
		mockedHTTPStatusCode int
	}{
		{
			name: "Version exists returns pass=true",
			spec: Spec{
				Name:    "requests",
				Version: "2.31.0",
				URL:     "https://pypi.example.com",
				Token:   "validtoken",
			},
			mockedBody:           existingPackageData,
			mockedURL:            "https://pypi.example.com/",
			mockedToken:          "validtoken",
			mockedHTTPStatusCode: 200,
			expectedPass:         true,
		},
		{
			name: "Version does not exist returns pass=false",
			spec: Spec{
				Name:    "requests",
				Version: "99.99.99",
				URL:     "https://pypi.example.com",
				Token:   "validtoken",
			},
			mockedBody:           existingPackageData,
			mockedURL:            "https://pypi.example.com/",
			mockedToken:          "validtoken",
			mockedHTTPStatusCode: 200,
			expectedPass:         false,
		},
		{
			name: "No version in spec uses source input",
			spec: Spec{
				Name:  "requests",
				URL:   "https://pypi.example.com",
				Token: "validtoken",
			},
			sourceInput:          "2.29.0",
			mockedBody:           existingPackageData,
			mockedURL:            "https://pypi.example.com/",
			mockedToken:          "validtoken",
			mockedHTTPStatusCode: 200,
			expectedPass:         true,
		},
		{
			name: "No version defined at all returns error",
			spec: Spec{
				Name:  "requests",
				URL:   "https://pypi.example.com",
				Token: "validtoken",
			},
			mockedBody:           existingPackageData,
			mockedURL:            "https://pypi.example.com/",
			mockedToken:          "validtoken",
			mockedHTTPStatusCode: 200,
			expectedError:        true,
		},
		{
			name: "Private registry with valid token returns correct result",
			spec: Spec{
				Name:    "requests",
				Version: "2.28.0",
				URL:     "https://pypi.example.com",
				Token:   "validtoken",
			},
			mockedBody:           existingPackageData,
			mockedURL:            "https://pypi.example.com/",
			mockedToken:          "validtoken",
			mockedHTTPStatusCode: 200,
			expectedPass:         true,
		},
		{
			name: "Yanked version is not considered available",
			spec: Spec{
				Name:    "requests",
				Version: "2.30.0",
				URL:     "https://pypi.example.com",
				Token:   "validtoken",
			},
			mockedBody:           existingPackageData,
			mockedURL:            "https://pypi.example.com/",
			mockedToken:          "validtoken",
			mockedHTTPStatusCode: 200,
			// 2.30.0 is yanked in existingPackageData
			expectedPass: false,
		},
		{
			name: "Pre-release PEP 440 version matches via original mapping",
			spec: Spec{
				Name:    "testpkg",
				Version: "1.0b2",
				URL:     "https://pypi.example.com",
				Token:   "validtoken",
				VersionFilter: version.Filter{
					Kind: "semver",
				},
			},
			mockedBody:           preReleasePackageData,
			mockedURL:            "https://pypi.example.com/",
			mockedToken:          "validtoken",
			mockedHTTPStatusCode: 200,
			// versions are normalized to semver but condition should match the original PEP 440 form
			expectedPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(tt.spec)
			require.NoError(t, err)

			p.webClient = GetMockClient(tt.mockedURL, tt.mockedToken, tt.mockedBody, tt.mockedHTTPStatusCode)

			pass, _, gotErr := p.Condition(context.Background(), tt.sourceInput, nil)

			if tt.expectedError {
				assert.Error(t, gotErr)
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, tt.expectedPass, pass)
		})
	}
}
