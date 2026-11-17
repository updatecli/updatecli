package npm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestSource(t *testing.T) {
	dir, err := CreateDummyRc()
	if err != nil {
		require.NoError(t, err)
	}
	defer os.RemoveAll(dir)

	// betweenAxiosReleases is a minimum age that the mocked "0.1.0" release passes and "0.2.0" doesn't
	betweenAxiosReleases := fmt.Sprintf("%dh", int(time.Since(time.Date(2014, time.September, 5, 0, 0, 0, 0, time.UTC)).Hours()))

	// undatedPackageData is a registry response reporting no release date for any version
	undatedPackageData := strings.Replace(existingPackageData, `,"0.1.0":"2014-08-29T23:08:36.810Z","0.2.0":"2014-09-12T20:06:33.167Z"`, "", 1)
	require.NotEqual(t, existingPackageData, undatedPackageData)
	// partiallyDatedPackageData is a registry response reporting no release date for "0.2.0"
	partiallyDatedPackageData := strings.Replace(existingPackageData, `,"0.2.0":"2014-09-12T20:06:33.167Z"`, "", 1)
	require.NotEqual(t, existingPackageData, partiallyDatedPackageData)

	tests := []struct {
		name                 string
		url                  string
		spec                 Spec
		expectedResult       string
		expectedError        bool
		expectedSkipped      bool
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
			name: "Skipped case of retrieving axios version with an unrealistic minimum age",
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
			expectedSkipped:      true,
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
			// "0.1.0" passes the minimum age but not the filter, "0.2.0" the filter but not the minimum age
			name: "Skipped case when every axios version matching the filter is too recent",
			spec: Spec{
				Name: "axios",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: ">=0.2.0",
				},
				Age:           age.Spec{Minimum: betweenAxiosReleases},
				URL:           "https://mycustomregistry.updatecli.io",
				RegistryToken: "mytoken",
			},
			mockedResponse:       true,
			mockedBody:           existingPackageData,
			mockedHTTPStatusCode: 200,
			mockedToken:          "mytoken",
			mockedUrl:            "https://mycustomregistry.updatecli.io",
			expectedSkipped:      true,
		},
		{
			name: "Failing case when the filter matches no axios version while every version is too recent",
			spec: Spec{
				Name: "axios",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: ">=9.0.0",
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
			// "0.1.0" matches the filter but is older than the maximum age, an age hold like a cooldown
			name: "Skipped case when every axios version matching the filter is above the maximum age",
			spec: Spec{
				Name: "axios",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "<0.2.0",
				},
				Age:           age.Spec{Minimum: "1h", Maximum: betweenAxiosReleases},
				URL:           "https://mycustomregistry.updatecli.io",
				RegistryToken: "mytoken",
			},
			mockedResponse:       true,
			mockedBody:           existingPackageData,
			mockedHTTPStatusCode: 200,
			mockedToken:          "mytoken",
			mockedUrl:            "https://mycustomregistry.updatecli.io",
			expectedSkipped:      true,
		},
		{
			name: "Skipped case of retrieving axios version with an unrealistic maximum age",
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
			expectedSkipped:      true,
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
			name: "Skipped case of retrieving axios version with an unrealistic minimum age and the default latest versionfilter",
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
			expectedSkipped:      true,
		},
		{
			// The npm autodiscovery sets a minimum age by default, which mustn't fail a registry without dates
			name: "Skipped case of a semver filter on a registry reporting no release date",
			spec: Spec{
				Name: "axios",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: ">=0.1.0",
				},
				Age:           age.Spec{Minimum: "3d"},
				URL:           "https://mycustomregistry.updatecli.io",
				RegistryToken: "mytoken",
			},
			mockedResponse:       true,
			mockedBody:           undatedPackageData,
			mockedHTTPStatusCode: 200,
			mockedToken:          "mytoken",
			mockedUrl:            "https://mycustomregistry.updatecli.io",
			expectedSkipped:      true,
		},
		{
			// "0.1.0" is too recent and "0.2.0", the only version matching the filter, has no release date
			name: "Skipped case of a semver filter matching only a version without release date",
			spec: Spec{
				Name: "axios",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: ">=0.2.0",
				},
				Age:           age.Spec{Minimum: "100y"},
				URL:           "https://mycustomregistry.updatecli.io",
				RegistryToken: "mytoken",
			},
			mockedResponse:       true,
			mockedBody:           partiallyDatedPackageData,
			mockedHTTPStatusCode: 200,
			mockedToken:          "mytoken",
			mockedUrl:            "https://mycustomregistry.updatecli.io",
			expectedSkipped:      true,
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
			err = got.Source(context.Background(), pathresolver.Resolver{}, &gotResult)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			// A cooldown discarding every published version skips the source, it doesn't fail it.
			if tt.expectedSkipped {
				assert.Equal(t, result.SKIPPED, gotResult.Result)
				return
			}

			assert.Equal(t, tt.expectedResult, gotResult.Information)
		})
	}

}
