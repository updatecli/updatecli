package gomod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCondition(t *testing.T) {
	tests := []struct {
		name             string
		spec             Spec
		expectedResult   bool
		expectedError    bool
		expectedErrorMsg error
	}{
		{
			spec: Spec{
				File:    "testdata/go.mod",
				Module:  "sigs.k8s.io/yaml",
				Version: "v1.3.0",
			},
			expectedResult: true,
		},
		{
			spec: Spec{
				File:    "testdata/go.mod",
				Module:  "sigs.k8s.io/yaml",
				Version: "v0.0.99",
			},
			expectedResult: false,
		},
		{
			spec: Spec{
				File:    "testdata/go.mod",
				Version: "v0.0.99",
			},
			expectedResult: false,
		},
		{
			name: "Test retrieving module from https",
			spec: Spec{
				File:    "https://raw.githubusercontent.com/updatecli/updatecli/v0.60.0/go.mod",
				Module:  "github.com/Masterminds/sprig/v3",
				Version: "v3.2.3",
			},
			expectedResult: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			require.NoError(t, err)
			gotResult, _, gotErr := got.Condition("", nil)
			if tt.expectedError {
				if assert.Error(t, gotErr) {
					assert.Equal(t, gotErr.Error(), tt.expectedErrorMsg.Error())
				}
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
