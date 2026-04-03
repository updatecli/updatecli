package pypi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestSource(t *testing.T) {
	tests := []struct {
		name                 string
		spec                 Spec
		expectedResult       string
		expectedError        bool
		mockedBody           string
		mockedURL            string
		mockedToken          string
		mockedHTTPStatusCode int
	}{
		{
			name: "Latest version retrieved when no filter set",
			spec: Spec{
				Name:  "requests",
				URL:   "https://pypi.example.com",
				Token: "validtoken",
			},
			mockedBody:           existingPackageData,
			mockedURL:            "https://pypi.example.com/",
			mockedToken:          "validtoken",
			mockedHTTPStatusCode: 200,
			expectedResult:       "2.31.0",
		},
		{
			name: "Semver filter >=2.29 returns highest matching non-yanked version",
			spec: Spec{
				Name:  "requests",
				URL:   "https://pypi.example.com",
				Token: "validtoken",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: ">=2.29",
				},
			},
			mockedBody:           existingPackageData,
			mockedURL:            "https://pypi.example.com/",
			mockedToken:          "validtoken",
			mockedHTTPStatusCode: 200,
			expectedResult:       "2.31.0",
		},
		{
			name: "Yanked versions excluded — semver filter only sees non-yanked",
			spec: Spec{
				Name:  "requests",
				URL:   "https://pypi.example.com",
				Token: "validtoken",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: ">=2.28",
				},
			},
			// 2.29, 2.30, 2.31 all yanked; only 2.28.0 available
			mockedBody:           yankedPackageData,
			mockedURL:            "https://pypi.example.com/",
			mockedToken:          "validtoken",
			mockedHTTPStatusCode: 200,
			expectedResult:       "2.28.0",
		},
		{
			name: "Non-existing package returns error",
			spec: Spec{
				Name:  "doesnotexist",
				URL:   "https://pypi.example.com",
				Token: "validtoken",
			},
			mockedBody:           nonExistingPackageData,
			mockedURL:            "https://pypi.example.com/",
			mockedToken:          "validtoken",
			mockedHTTPStatusCode: 404,
			expectedError:        true,
		},
		{
			name: "Private registry with bad token returns error",
			spec: Spec{
				Name:  "requests",
				URL:   "https://pypi.example.com",
				Token: "wrongtoken",
			},
			mockedBody:           existingPackageData,
			mockedURL:            "https://pypi.example.com/",
			mockedToken:          "validtoken",
			mockedHTTPStatusCode: 200,
			expectedError:        true,
		},
		{
			name: "Semver exact constraint returns matching version",
			spec: Spec{
				Name:  "requests",
				URL:   "https://pypi.example.com",
				Token: "validtoken",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "2.29.0",
				},
			},
			mockedBody:           existingPackageData,
			mockedURL:            "https://pypi.example.com/",
			mockedToken:          "validtoken",
			mockedHTTPStatusCode: 200,
			expectedResult:       "2.29.0",
		},
		{
			name: "PEP 440 filter returns raw versions without normalization",
			spec: Spec{
				Name:  "testpkg",
				URL:   "https://pypi.example.com",
				Token: "validtoken",
				VersionFilter: version.Filter{
					Kind:    "pep440",
					Pattern: ">=1.0a1",
				},
			},
			mockedBody:           preReleasePackageData,
			mockedURL:            "https://pypi.example.com/",
			mockedToken:          "validtoken",
			mockedHTTPStatusCode: 200,
			expectedResult:       "1.0b2",
		},
		{
			name: "Latest version is yanked returns error",
			spec: Spec{
				Name:  "requests",
				URL:   "https://pypi.example.com",
				Token: "validtoken",
			},
			mockedBody:           yankedLatestPackageData,
			mockedURL:            "https://pypi.example.com/",
			mockedToken:          "validtoken",
			mockedHTTPStatusCode: 200,
			expectedError:        true,
		},
		{
			name: "Post release normalized to base with semver filter",
			spec: Spec{
				Name:  "testpkg",
				URL:   "https://pypi.example.com",
				Token: "validtoken",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: ">=1.0.0",
				},
			},
			mockedBody:           postReleasePackageData,
			mockedURL:            "https://pypi.example.com/",
			mockedToken:          "validtoken",
			mockedHTTPStatusCode: 200,
			// 1.0.post1 normalizes to 1.0.0, which is >= 1.0.0; original PEP 440 form returned
			expectedResult: "1.0.post1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := New(tt.spec)
			require.NoError(t, err)

			p.webClient = GetMockClient(tt.mockedURL, tt.mockedToken, tt.mockedBody, tt.mockedHTTPStatusCode)

			gotResult := result.Source{}
			err = p.Source(context.Background(), "", &gotResult)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, gotResult.Information)
		})
	}
}
