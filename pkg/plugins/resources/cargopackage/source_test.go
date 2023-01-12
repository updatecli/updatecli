package cargopackage

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestSource(t *testing.T) {
	dir, err := CreateDummyIndex()
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir)
	tests := []struct {
		name           string
		url            string
		spec           Spec
		expectedResult string
		expectedError  bool
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
				IndexDir: dir,
				Package:  "crate-test",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~0.1",
				},
			},
			expectedResult: "0.1.0",
			expectedError:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec, false)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			gotVersion, err := got.Source("")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, gotVersion)
		})
	}

}
