package toolversions

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCondition(t *testing.T) {

	testData := []struct {
		name             string
		spec             Spec
		expectedResult   bool
		expectedErrorMsg error
		wantErr          bool
	}{
		{
			name: "Default scenario",
			spec: Spec{
				File:  "testdata/.tool-versions",
				Key:   "bats",
				Value: "1.0.0",
			},
			expectedResult: true,
		},
		{
			name: "Test key do not exist with empty result",
			spec: Spec{
				File:  "testdata/.tool-versions",
				Key:   "empty",
				Value: "",
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find value for key \"empty\" from file \"testdata/.tool-versions\""),
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/.tool-versions",
				Key:   "doNotExist",
				Value: "",
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find value for key \"doNotExist\" from file \"testdata/.tool-versions\""),
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			toml, err := New(tt.spec)

			require.NoError(t, err)

			gotResult, _, gotErr := toml.Condition("", nil)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), gotErr.Error())
			} else {
				require.NoError(t, gotErr)
			}

			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
