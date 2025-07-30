package json

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
				File:  "testdata/data.json",
				Key:   ".firstName",
				Value: "Jack",
			},
			expectedResult: true,
		},
		{
			name: "Default scenario with Dasel v2",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    ".firstName",
				Value:  "Jack",
				Engine: strPtr(ENGINEDASEL_V2),
			},
			expectedResult: true,
		},
		{
			name: "Multiple key scenario",
			spec: Spec{
				File:  "testdata/data.json",
				Query: "phoneNumbers.[0].type",
				Value: "home",
			},
			expectedResult: true,
		},
		{
			name: "Multiple key scenario",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    "phoneNumbers.[0].type",
				Value:  "home",
				Engine: strPtr(ENGINEDASEL_V2),
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
				File:  "testdata/data.json",
				Query: ".doNotExist",
				Value: "",
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

			got, _, gotErr := j.Condition("", nil)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), gotErr.Error())
			} else {
				require.NoError(t, gotErr)
			}

			assert.Equal(t, tt.expectedResult, got)
		})
	}
}
