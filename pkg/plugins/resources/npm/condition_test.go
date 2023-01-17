package npm

import (
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCondition(t *testing.T) {
	dir, err := CreateDummyRc()
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
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
		mockedHTTPStatusCode int
	}{
		{
			name: "Passing case of retrieving axios versions ",
			spec: Spec{
				Name:    "axios",
				Version: "1.0.0",
			},
			expectedResult: true,
			expectedError:  false,
		},
		{
			name: "Passing case of retrieving latest axios version using latest rule ",
			spec: Spec{
				Name:    "axios",
				Version: "99.99.99",
			},
			expectedResult: false,
			expectedError:  false,
		},
		{
			name: "Passing case of retrieving axios versions using custom default registry",
			spec: Spec{
				Name:          "axios",
				Version:       "0.2.0",
				URL:           "https://mycustomregistry.updatecli.io",
				RegistryToken: "mytoken",
			},
			expectedResult:       true,
			expectedError:        false,
			mockedResponse:       true,
			mockedBody:           existingPackageData,
			mockedHTTPStatusCode: 200,
			mockedToken:          "mytoken",
			mockedUrl:            "https://mycustomregistry.updatecli.io",
		},
		{
			name: "Passing case of retrieving latest axios version using latest rule using custom default registry ",
			spec: Spec{
				Name:          "axios",
				Version:       "99.99.99",
				URL:           "https://mycustomregistry.updatecli.io",
				RegistryToken: "mytoken",
			},
			expectedResult:       false,
			expectedError:        false,
			mockedResponse:       true,
			mockedBody:           existingPackageData,
			mockedHTTPStatusCode: 200,
			mockedToken:          "mytoken",
			mockedUrl:            "https://mycustomregistry.updatecli.io",
		},
		{
			name: "Passing case of retrieving axios versions using private registry in npmrc",
			spec: Spec{
				Name:      "@TestScope/test",
				Version:   "0.2.0",
				NpmrcPath: filepath.Join(dir, ".npmrc"),
			},
			expectedResult:       true,
			expectedError:        false,
			mockedResponse:       true,
			mockedBody:           existingScopedPackageData,
			mockedHTTPStatusCode: 200,
			mockedToken:          "mytoken",
			mockedUrl:            "https://mycustomregistry.updatecli.io",
		},
		{
			name: "Passing case of retrieving latest axios version using latest rule using private registry in npmrc ",
			spec: Spec{
				Name:      "@TestScope/test",
				Version:   "99.99.99",
				NpmrcPath: filepath.Join(dir, ".npmrc"),
			},
			expectedResult:       false,
			expectedError:        false,
			mockedResponse:       true,
			mockedBody:           existingScopedPackageData,
			mockedHTTPStatusCode: 200,
			mockedToken:          "mytoken",
			mockedUrl:            "https://mycustomregistry.updatecli.io",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.mockedResponse {
				got.webClient = GetMockClient(tt.mockedUrl, tt.mockedToken, tt.mockedBody, tt.mockedHTTPStatusCode)
			}
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			gotVersion, err := got.Condition("")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, gotVersion)
		})
	}

}
