package csv

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestTarget(t *testing.T) {

	testData := []struct {
		name             string
		spec             Spec
		sourceInput      string
		expectedResult   bool
		expectedErrorMsg error
		wantErr          bool
	}{
		{
			name: "Default successful workflow",
			spec: Spec{
				File: "testdata/data.csv",
				Key:  ".[0].firstname",
			},
			sourceInput:    "Tom",
			expectedResult: true,
		},
		{
			name: "Default successful workflow",
			spec: Spec{
				File:  "testdata/data.2.csv",
				Key:   ".[0].firstname",
				Comma: ';',
			},
			sourceInput:    "Tom",
			expectedResult: true,
		},
		{
			name: "Do not exist query workflow",
			spec: Spec{
				File:  "testdata/data.2.csv",
				Key:   ".[0].DoNotExist",
				Comma: ';',
			},
			sourceInput:      "Tom",
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find value for query \".[0].DoNotExist\" from file \"testdata/data.2.csv\""),
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			c, err := New(tt.spec)

			require.NoError(t, err)

			gotResult, err := c.Target(tt.sourceInput, true)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
