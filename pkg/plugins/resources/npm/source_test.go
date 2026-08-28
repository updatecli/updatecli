package npm

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
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
		expectedNewError     bool
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
					Pattern: "~0.27",
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
		{
			name: "Passing case of retrieving axios version with a minimum age",
			spec: Spec{
				Name: "axios",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~0",
				},
				Age:           age.Spec{Minimum: "1y"},
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
			name: "Failing case of retrieving axios version with an unrealistic minimum age",
			spec: Spec{
				Name: "axios",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~0",
				},
				Age:           age.Spec{Minimum: "100y"},
				URL:           "https://mycustomregistry.updatecli.io",
				RegistryToken: "mytoken",
			},
			mockedResponse:       true,
			mockedBody:           existingPackageData,
			mockedHTTPStatusCode: 200,
			mockedToken:          "mytoken",
			mockedUrl:            "https://mycustomregistry.updatecli.io",
			expectedError:        true,
		},
		{
			name: "Passing case of retrieving axios version with a maximum age",
			spec: Spec{
				Name: "axios",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~0",
				},
				Age:           age.Spec{Maximum: "100y"},
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
			name: "Failing case of retrieving axios version with an unrealistic maximum age",
			spec: Spec{
				Name: "axios",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~0",
				},
				Age:           age.Spec{Maximum: "1s"},
				URL:           "https://mycustomregistry.updatecli.io",
				RegistryToken: "mytoken",
			},
			mockedResponse:       true,
			mockedBody:           existingPackageData,
			mockedHTTPStatusCode: 200,
			mockedToken:          "mytoken",
			mockedUrl:            "https://mycustomregistry.updatecli.io",
			expectedError:        true,
		},
		{
			name: "Passing case of retrieving axios version with a minimum age and the default latest versionfilter",
			spec: Spec{
				Name:          "axios",
				Age:           age.Spec{Minimum: "1y"},
				URL:           "https://mycustomregistry.updatecli.io",
				RegistryToken: "mytoken",
			},
			mockedResponse:       true,
			mockedBody:           existingPackageData,
			mockedHTTPStatusCode: 200,
			mockedToken:          "mytoken",
			mockedUrl:            "https://mycustomregistry.updatecli.io",
			// dist-tags latest is 0.1.0 and it matches the age filter
			expectedResult: "0.1.0",
		},
		{
			name: "Failing case of retrieving axios version with an unrealistic minimum age and the default latest versionfilter",
			spec: Spec{
				Name:          "axios",
				Age:           age.Spec{Minimum: "100y"},
				URL:           "https://mycustomregistry.updatecli.io",
				RegistryToken: "mytoken",
			},
			mockedResponse:       true,
			mockedBody:           existingPackageData,
			mockedHTTPStatusCode: 200,
			mockedToken:          "mytoken",
			mockedUrl:            "https://mycustomregistry.updatecli.io",
			expectedError:        true,
		},
		{
			name: "Failing case of an invalid age spec",
			spec: Spec{
				Name: "axios",
				Age:  age.Spec{Minimum: "notaduration"},
			},
			expectedNewError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			if tt.expectedNewError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.mockedResponse {
				got.webClient = GetMockClient(tt.mockedUrl, tt.mockedToken, tt.mockedBody, tt.mockedHTTPStatusCode)
			}
			gotResult := result.Source{}
			err = got.Source(context.Background(), utils.Resolver{}, &gotResult)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, gotResult.Information)
		})
	}

}
