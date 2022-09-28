package json

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
			name: "Default scenario",
			spec: Spec{
				File:  "testdata/data.json",
				Key:   ".firstName",
				Value: "Jack",
			},
			expectedResult: true,
		},
		{
			name: "Multiple key scenario",
			spec: Spec{
				File:     "testdata/data.json",
				Key:      "phoneNumbers.[0].type",
				Value:    "home",
				Multiple: true,
			},
			expectedResult: true,
		},
		// Doesn't seem to be working on Dasel side.
		//{
		//	name: "none multiple key scenario",
		//	spec: Spec{
		//		File:  "testdata/data.json",
		//		Key:   "phoneNumbers.[*].type",
		//		Value: "another home",
		//	},
		//	expectedResult: true,
		//},
		{
			name: "Default successful workflow with empty result",
			spec: Spec{
				File:  "testdata/data.json",
				Key:   ".surname",
				Value: "",
			},
			expectedResult: true,
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/data.json",
				Key:   ".doNotExist",
				Value: "",
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find value for query \".doNotExist\" from file \"testdata/data.json\""),
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:     "testdata/data.json",
				Key:      ".doNotExist",
				Value:    "",
				Multiple: true,
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find multiple value for query \".doNotExist\" from file \"testdata/data.json\""),
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			j, err := New(tt.spec)

			require.NoError(t, err)

			gotResult, err := j.Condition("")

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
