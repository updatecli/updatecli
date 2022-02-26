package jsonschema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type Dataset []Data

// Data represent a single json schema test data
type Data struct {
	// name defines a test name
	name                  string
	updatecliManifestPath string
	expectedErrMessage    error
	expectedErr           bool
	expectedResult        bool
}

var (
	TestDataset Dataset = Dataset{
		{
			name:                  "Test using e2e updateli manifest",
			updatecliManifestPath: "../../../../e2e/updatecli.d",
			expectedResult:        true,
			expectedErrMessage:    nil,
			expectedErr:           true,
		},
	}
)

// TestValidate tests the function Validate by Using updatecli manifest
// from the e2e directory.
func TestValidate(t *testing.T) {

	for _, d := range TestDataset {
		t.Run(d.name, func(t *testing.T) {
			gotResult, gotErr := Validate(d.updatecliManifestPath)
			if d.expectedErr {
				require.Error(t, gotErr)
			}

			if gotResult != d.expectedResult {
				t.Error()
			}

			require.NoError(t, gotErr)
		})
	}
}
