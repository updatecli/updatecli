package gomodule

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
				Path:    "github.com/updatecli/updatecli",
				Version: "v0.47.2",
			},
			expectedResult: true,
		},
		{
			name: "Test go module with upper case character",
			spec: Spec{
				Path:    "github.com/MakeNowJust/heredoc",
				Version: "v1.0.0",
			},
			expectedResult: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			require.NoError(t, err)
			gotVersion, err := got.Condition("")
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, gotVersion)
		})
	}

}
