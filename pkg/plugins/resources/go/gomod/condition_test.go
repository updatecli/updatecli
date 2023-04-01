package gomod

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
