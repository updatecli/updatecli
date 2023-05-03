package language

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
				Version: "1.20.0",
			},
			expectedResult: true,
		},
		{
			name: "test that it doesn't exist",
			spec: Spec{
				Version: "1.11.1111",
			},
			expectedResult:   false,
			expectedError:    true,
			expectedErrorMsg: errors.New("golang version \"1.11.1111\" doesn't exist"),
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
