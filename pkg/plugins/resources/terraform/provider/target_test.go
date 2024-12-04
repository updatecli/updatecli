package provider

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestTarget(t *testing.T) {
	testData := []struct {
		name             string
		spec             Spec
		sourceInput      string
		expectedResult   bool
		expectedErrorMsg error
		wantErr          bool
	}{
		{
			name: "Success - No change",
			spec: Spec{
				File:     "testdata/versions.tf",
				Provider: "azurerm",
			},
			sourceInput:    "3.69.0",
			expectedResult: false,
		},
		{
			name: "Success - Expected change",
			spec: Spec{
				File:     "testdata/versions.tf",
				Provider: "azurerm",
			},
			sourceInput:    "3.70.0",
			expectedResult: true,
		},
		{
			name: "Success - Expected change using Value",
			spec: Spec{
				File:     "testdata/versions.tf",
				Provider: "azurerm",
				Value:    "3.70.0",
			},
			expectedResult: true,
		},
		{
			name: "Failure - File does not exists",
			spec: Spec{
				File:     "testdata/doNotExist.hcl",
				Provider: "azurerm",
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New(`✗ The specified file "testdata/doNotExist.hcl" does not exist`),
		},
		{
			name: "Failure - Path does not exists",
			spec: Spec{
				File:     "testdata/versions.tf",
				Provider: "null",
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New(`✗ cannot find value for "null" from file "testdata/versions.tf"`),
		},
		{
			name: "Failure - HTTP Target",
			spec: Spec{
				File:     "http://localhost/doNotExist.hcl",
				Provider: "azurerm",
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New(`✗ URL scheme is not supported for HCL target: "http://localhost/doNotExist.hcl"`),
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			l, err := New(tt.spec)

			require.NoError(t, err)

			gotResult := result.Target{}
			err = l.Target(result.SourceInformation{Value: tt.sourceInput}, nil, true, &gotResult)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Changed)
		})
	}
}
