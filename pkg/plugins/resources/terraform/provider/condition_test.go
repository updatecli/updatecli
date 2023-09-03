package provider

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestCondition(t *testing.T) {
	testData := []struct {
		name             string
		spec             Spec
		source           string
		expectedResult   bool
		expectedErrorMsg error
		wantErr          bool
	}{
		{
			name: "Success - Using Value",
			spec: Spec{
				File:     "testdata/versions.tf",
				Provider: "azurerm",
				Value:    "3.69.0",
			},
			source:         "",
			expectedResult: true,
		},
		{
			name: "Success - Using Source",
			spec: Spec{
				File:     "testdata/versions.tf",
				Provider: "azurerm",
			},
			source:         "3.69.0",
			expectedResult: true,
		},
		{
			name: "Failure - Using Source",
			spec: Spec{
				File:     "testdata/versions.tf",
				Provider: "azurerm",
			},
			source:         "3.70.0",
			expectedResult: false,
		},
		{
			name: "Failure - File does not exists",
			spec: Spec{
				File:     "testdata/doNotExist.hcl",
				Provider: "azurerm",
			},
			source:           "",
			wantErr:          true,
			expectedErrorMsg: errors.New(`✗ The specified file "testdata/doNotExist.hcl" does not exist`),
		},
		{
			name: "Failure - Path does not exists",
			spec: Spec{
				File:     "testdata/versions.tf",
				Provider: "null",
			},
			source:           "",
			wantErr:          true,
			expectedErrorMsg: errors.New(`✗ cannot find value for "null" from file "testdata/versions.tf"`),
		},
		{
			name: "Failure - Multiple Files",
			spec: Spec{
				Files:    []string{"testdata/data.hcl", "testdata/data2.hcl"},
				Provider: "azurerm",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("✗ terraform/lock condition only supports one file"),
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			l, err := New(tt.spec)

			require.NoError(t, err)

			gotResult := result.Condition{}
			err = l.Condition(tt.source, nil, &gotResult)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Pass)
		})
	}
}
