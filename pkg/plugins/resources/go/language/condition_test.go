package language

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCondition(t *testing.T) {
	tests := []struct {
		name           string
		spec           Spec
		expectedResult bool
		expectedError  bool
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
			foundVersion, err := got.Condition("")
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, foundVersion)
		})
	}

}
