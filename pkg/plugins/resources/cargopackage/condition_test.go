package cargopackage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCondition(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		spec           Spec
		expectedResult bool
		expectedError  bool
	}{
		{
			name: "Passing case of retrieving rand versions ",
			spec: Spec{
				Package: "rand",
				Version: "0.7.3",
			},
			expectedResult: true,
			expectedError:  false,
		},
		{
			name: "Passing case of retrieving latest rand version using latest rule ",
			spec: Spec{
				Package: "rand",
				Version: "99.99.99",
			},
			expectedResult: false,
			expectedError:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			gotVersion, err := got.Condition("")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, gotVersion)
		})
	}
}
