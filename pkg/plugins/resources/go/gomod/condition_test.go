package gomod

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
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
			expectedResult:   false,
			expectedError:    true,
			expectedErrorMsg: errors.New("golang module version \"v1.3.0\" found for \"sigs.k8s.io/yaml\", expecting \"v0.0.99\""),
		},
		{
			spec: Spec{
				File:    "testdata/go.mod",
				Version: "v0.0.99",
			},
			expectedResult:   false,
			expectedError:    true,
			expectedErrorMsg: errors.New("golang version \"1.20\" found, expecting \"v0.0.99\""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			require.NoError(t, err)
			gotResult := result.Condition{}
			err = got.Condition("", nil, &gotResult)
			if tt.expectedError {
				if assert.Error(t, err) {
					assert.Equal(t, err.Error(), tt.expectedErrorMsg.Error())
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, gotResult.Pass)
		})
	}
}
