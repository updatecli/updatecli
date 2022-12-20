package csv

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
				File:  "testdata/data.csv",
				Key:   ".[0].firstname",
				Value: "John",
			},
			expectedResult: true,
		},
		{
			name: "Deprecated Multiple query scenario",
			spec: Spec{
				File:     "testdata/data.csv",
				Key:      ".[*].firstname",
				Value:    "John",
				Multiple: true,
			},
			expectedResult: false,
		},
		{
			name: "Query scenario",
			spec: Spec{
				File:  "testdata/data.csv",
				Query: ".[*].firstname",
				Value: "John",
			},
			expectedResult: false,
		},
		{
			name: "Multiple query scenario",
			spec: Spec{
				File:  "testdata/data.csv",
				Query: ".[*].surname",
				Value: "",
			},
			expectedResult: false,
		},
		{
			name: "Default scenario 2",
			spec: Spec{
				File:  "testdata/data.2.csv",
				Key:   ".[0].firstname",
				Comma: ';',
				Value: "John",
			},
			expectedResult: true,
		},
		{
			name: "Default successful workflow with empty result",
			spec: Spec{
				File:  "testdata/data.csv",
				Key:   ".[0].surname",
				Value: "",
			},
			expectedResult: true,
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/data.csv",
				Key:   ".doNotExist",
				Value: "",
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find value for query \".doNotExist\" from file \"testdata/data.csv\""),
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			c, err := New(tt.spec)

			require.NoError(t, err)

			gotResult, err := c.Condition("")

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
