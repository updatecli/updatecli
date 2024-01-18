package language

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
				Version: "1.20.0",
			},
			expectedResult: true,
		},
		{
			name: "test that it doesn't exist",
			spec: Spec{
				Version: "1.11.1111",
			},
			expectedResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			require.NoError(t, err)

			gotResult, _, err := got.Condition("", nil)
			if tt.expectedError {
				if assert.Error(t, err) {
					assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}

}
