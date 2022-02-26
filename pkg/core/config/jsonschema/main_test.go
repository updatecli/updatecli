package jsonschema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type Dataset []Data

// Data represent a single json schema test data
type Data struct {
	// name defines a test name
	name           string
	config         Config
	expectedErr    bool
	expectedResult bool
}

var (
	TestDataset Dataset = Dataset{
		{
			name: "Test using e2e updateli manifest, jenkins.yaml",
			config: Config{
				MainJsonSchema:         "../../../../schema/config.json",
				UpdatecliConfiguration: "../../../../e2e/updatecli.d/jenkins.yaml",
			},
			expectedResult: true,
			expectedErr:    false,
		},
		{
			name: "Test using e2e updateli manifest",
			config: Config{
				MainJsonSchema:         "../../../../schema/config.json",
				UpdatecliConfiguration: "../../../../e2e/updatecli.d",
			},
			expectedResult: true,
			expectedErr:    true,
		},
	}
)

// TestValidate tests the function Validate by Using updatecli manifest
// from the e2e directory.
func TestValidate(t *testing.T) {

	for _, d := range TestDataset {
		t.Run(d.name, func(t *testing.T) {
			gotResult, gotErr := d.config.Validate()
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
