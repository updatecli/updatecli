package cargopackage

import (
	"os"
	"testing"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/cargo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestSource(t *testing.T) {
	dir, err := CreateDummyIndex()
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
		mockedHeaderFormat   string
		mockedHTTPStatusCode int
	}{
		{
			name: "Passing case of retrieving rand version from the default index api",
			spec: Spec{
				Package: "rand",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~0.7",
				},
			},
			expectedResult: "0.7.3",
			expectedError:  false,
		},
		{
			name: "Passing case of retrieving crate-test version from the filesystem index",
			spec: Spec{
				Registry: cargo.Registry{
					RootDir: dir,
				},
				Package: "crate-test",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~0.1",
				},
			},
			expectedResult: "0.1.0",
			expectedError:  false,
		},
		{
			name: "Passing case of retrieving crate-test version from a mocked private registry",
			spec: Spec{
				Registry: cargo.Registry{
					URL: "https://crates.io/api/v1/crates",
					Auth: cargo.InlineKeyChain{
						Token:        "mytoken",
						HeaderFormat: "Bearer %s",
					},
				},
				Package: "crate-test",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~0.1",
				},
			},
			expectedResult:       "0.1.0",
			expectedError:        false,
			mockedResponse:       true,
			mockedBody:           existingPackageData,
			mockedUrl:            "https://crates.io/api/v1/crates",
			mockedToken:          "mytoken",
			mockedHeaderFormat:   "Bearer %s",
			mockedHTTPStatusCode: existingPackageStatus,
		},
		{
			name: "Failing case of retrieving nonexistent package from a mocked private registry",
			spec: Spec{
				Registry: cargo.Registry{
					URL: "https://crates.io/api/v1/crates",
					Auth: cargo.InlineKeyChain{
						Token:        "mytoken",
						HeaderFormat: "Bearer %s",
					},
				},
				Package: "crate-test-non-existing",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~0.1",
				},
			},
			expectedError:        true,
			mockedResponse:       true,
			mockedBody:           nonExistingPackageData,
			mockedUrl:            "https://crates.io/api/v1/crates",
			mockedToken:          "mytoken",
			mockedHeaderFormat:   "Bearer %s",
			mockedHTTPStatusCode: nonExistingPackageStatus,
		},
		{
			name: "Failing case of retrieving existing package from a mocked private registry but with bad auth",
			spec: Spec{
				Registry: cargo.Registry{
					URL: "https://crates.io/api/v1/crates",
					Auth: cargo.InlineKeyChain{
						Token:        "bad token",
						HeaderFormat: "Bearer %s",
					},
				},
				Package: "crate-test",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~0.1",
				},
			},
			expectedError:        true,
			mockedResponse:       true,
			mockedBody:           existingPackageData,
			mockedUrl:            "https://crates.io/api/v1/crates",
			mockedToken:          "mytoken",
			mockedHeaderFormat:   "Bearer %s",
			mockedHTTPStatusCode: existingPackageStatus,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec, false)
			require.NoError(t, err)
			if tt.mockedResponse {
				got.webClient = GetMockClient(tt.mockedUrl, tt.mockedToken, tt.mockedBody, tt.mockedHTTPStatusCode, tt.mockedHeaderFormat)
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
