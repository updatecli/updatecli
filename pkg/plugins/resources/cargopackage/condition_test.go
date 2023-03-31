package cargopackage

import (
	"os"
	"testing"

	"github.com/updatecli/updatecli/pkg/plugins/utils/cargo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCondition(t *testing.T) {
	dir, err := CreateDummyIndex()
	defer os.RemoveAll(dir)
	if err != nil {
		require.NoError(t, err)
	}

	tests := []struct {
		name                 string
		url                  string
		spec                 Spec
		expectedResult       bool
		expectedError        bool
		mockedResponse       bool
		mockedBody           string
		mockedUrl            string
		mockedToken          string
		mockedHeaderFormat   string
		mockedHTTPStatusCode int
	}{
		{
			name: "Retrieving existing rand version from the default index api",
			spec: Spec{
				Package: "rand",
				Version: "0.7.2",
			},
			expectedResult: true,
			expectedError:  false,
		},
		{
			name: "Retrieving non-existing rand version from the default index api",
			spec: Spec{
				Package: "rand",
				Version: "99.99.99",
			},
			expectedResult: false,
			expectedError:  false,
		},
		{
			name: "Retrieving existing crate-test version from the filesystem index",
			spec: Spec{
				Registry: cargo.Registry{
					RootDir: dir,
				},
				Package: "crate-test",
				Version: "0.2.2",
			},
			expectedResult: true,
			expectedError:  false,
		},
		{
			name: "Retrieving existing yanked crate-test version from the filesystem index",
			spec: Spec{
				Registry: cargo.Registry{
					RootDir: dir,
				},
				Package: "crate-test",
				Version: "0.2.3",
			},
			expectedResult: false,
			expectedError:  false,
		},
		{
			name: "Retrieving non-existing yanked crate-test version from the filesystem index",
			spec: Spec{
				Registry: cargo.Registry{
					RootDir: dir,
				},
				Package: "crate-test",
				Version: "99.99.99",
			},
			expectedResult: false,
			expectedError:  false,
		},
		{
			name: "Retrieving existing crate-test version from the mocked private registry",
			spec: Spec{
				Registry: cargo.Registry{
					URL: "https://crates.io/api/v1/crates",
					Auth: cargo.InlineKeyChain{
						Token:        "mytoken",
						HeaderFormat: "Bearer %s",
					},
				},
				Package: "crate-test",
				Version: "0.2.0",
			},
			expectedResult:       true,
			expectedError:        false,
			mockedResponse:       true,
			mockedBody:           existingPackageData,
			mockedUrl:            "https://crates.io/api/v1/crates",
			mockedToken:          "mytoken",
			mockedHeaderFormat:   "Bearer %s",
			mockedHTTPStatusCode: existingPackageStatus,
		},
		{
			name: "Retrieving non-existing crate-test version from the mocked private registry",
			spec: Spec{
				Registry: cargo.Registry{
					URL: "https://crates.io/api/v1/crates",
					Auth: cargo.InlineKeyChain{
						Token:        "mytoken",
						HeaderFormat: "Bearer %s",
					},
				},
				Package: "crate-test",
				Version: "99.99.99",
			},
			expectedResult:       false,
			expectedError:        false,
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
			gotVersion, err := got.Condition("")
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, gotVersion)
		})
	}
}
