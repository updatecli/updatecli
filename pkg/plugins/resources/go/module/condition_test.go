package gomodule

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
			expectedResult:   true,
			expectedError:    true,
			expectedErrorMsg: errors.New("version \"v0.0.0\" doesn't exist"),
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
					assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, gotResult.Pass)
		})
	}

}
