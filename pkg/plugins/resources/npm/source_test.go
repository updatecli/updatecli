package npm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestSource(t *testing.T) {
	dir, err := CreateDummyRc()
	if err != nil {
		require.NoError(t, err)
	}
	defer os.RemoveAll(dir)
	tests := []struct {
		name                 string
		url                  string
		spec                 Spec
		expectedResult       string
		expectedError        bool
		mockedResponse       bool
		mockedBody           string
		mockedUrl            string
		mockedToken          string
		mockedHTTPStatusCode int
	}{
		{
			name: "Passing case of retrieving axios versions ",
			spec: Spec{
				Name: "axios",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~0",
				},
			},
			expectedResult: "0.27.2",
			expectedError:  false,
		},
		{
			name: "Passing case of retrieving latest axios version using private registry",
			spec: Spec{
				Name: "axios",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~0",
				},
				URL:           "https://mycustomregistry.updatecli.io",
				RegistryToken: "mytoken",
			},
			mockedResponse:       true,
			mockedBody:           existingPackageData,
			mockedHTTPStatusCode: 200,
			mockedToken:          "mytoken",
			mockedUrl:            "https://mycustomregistry.updatecli.io",
			expectedResult:       "0.2.0",
		},
		{
			name: "Failing case of retrieving latest axios version using private registry but bad token",
			spec: Spec{
				Name: "axios",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~0",
				},
				URL:           "https://mycustomregistry.updatecli.io",
				RegistryToken: "badtoken",
			},
			mockedResponse:       true,
			mockedBody:           existingPackageData,
			mockedHTTPStatusCode: 200,
			mockedToken:          "mytoken",
			mockedUrl:            "https://mycustomregistry.updatecli.io",
			expectedError:        true,
		},
		{
			name: "Failing case of retrieving latest nonexistent package using private registry",
			spec: Spec{
				Name: "axiosnonexistent",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~0",
				},
				URL:           "https://mycustomregistry.updatecli.io",
				RegistryToken: "mytoken",
			},
			mockedResponse:       true,
			mockedBody:           nonExistingPackageData,
			mockedHTTPStatusCode: 404,
			mockedToken:          "mytoken",
			mockedUrl:            "https://mycustomregistry.updatecli.io",
			expectedError:        true,
		},
		{
			name: "Passing case of retrieving latest @TestScope:registry version using private registry in npmrc",
			spec: Spec{
				Name: "@TestScope/test",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~0",
				},
				NpmrcPath: filepath.Join(dir, ".npmrc"),
			},
			mockedResponse:       true,
			mockedBody:           existingScopedPackageData,
			mockedHTTPStatusCode: 200,
			mockedToken:          "mytoken",
			mockedUrl:            "https://mycustomregistry.updatecli.io",
			expectedResult:       "0.2.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			require.NoError(t, err)
			if tt.mockedResponse {
				got.webClient = GetMockClient(tt.mockedUrl, tt.mockedToken, tt.mockedBody, tt.mockedHTTPStatusCode)
			}
			gotResult := result.Source{}
			err = got.Source("", &gotResult)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, gotResult.Information)
		})
	}

}
