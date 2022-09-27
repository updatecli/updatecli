package toml

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
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
			name: "Default successful multiple update workflow",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  ".employees.[*].role",
			},
			expectedResult: false,
		},
		{
			name: "Successful conditional multiple update workflow",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  ".employees.(address=AU).role",
			},
			expectedResult: false,
		},
		{
			name: "Successful multiple map update workflow",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  ".benefits.[0].country.(country=UK).name",
			},
			expectedResult: false,
		},
		{
			name: "Default scenario",
			spec: Spec{
				File:  "testdata/data.toml",
				Key:   ".owner.firstName",
				Value: "Jack",
			},
			expectedResult: true,
		},
		{
			name: "Default successful workflow with empty result",
			spec: Spec{
				File:  "testdata/data.toml",
				Key:   ".owner.surname",
				Value: "",
			},
			expectedResult: true,
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/data.toml",
				Key:   ".doNotExist",
				Value: "",
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find multiple value for query \".doNotExist\" from file \"testdata/data.toml\""),
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			toml, err := New(tt.spec)

			require.NoError(t, err)

			gotResult, err := toml.Condition("")

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
