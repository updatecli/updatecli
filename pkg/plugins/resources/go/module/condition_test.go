package gomodule

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
			name: "canonical test",
			spec: Spec{
				Module:  "github.com/updatecli/updatecli",
				Version: "v0.47.2",
			},
			expectedResult: true,
		},
		{
			name: "Test go module with upper case character",
			spec: Spec{
				Module:  "github.com/MakeNowJust/heredoc",
				Version: "v1.0.0",
			},
			expectedResult: true,
		},
		{
			name: "Test go module version do not exist",
			spec: Spec{
				Module:  "github.com/MakeNowJust/heredoc",
				Version: "v0.0.0",
			},
			expectedResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			require.NoError(t, err)
			gotResult, _, gotErr := got.Condition("", nil)
			if tt.expectedError {
				if assert.Error(t, gotErr) {
					assert.Equal(t, tt.expectedErrorMsg.Error(), gotErr.Error())
				}
				return
			}
			require.NoError(t, gotErr)
			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}

}
