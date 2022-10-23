package json

import (
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
				File: "testdata/data.json",
				Key:  ".firstName",
			},
			sourceInput:    "Jack",
			expectedResult: false,
		},
		{
			name: "Default successful workflow",
			spec: Spec{
				File: "testdata/data.json",
				Key:  ".firstName",
			},
			sourceInput:    "Tom",
			expectedResult: true,
		},
		{
			name: "Default successful workflow",
			spec: Spec{
				File:  "testdata/data.json",
				Query: ".phoneNumbers.[*].type",
			},
			sourceInput:    "Tom",
			expectedResult: true,
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			j, err := New(tt.spec)

			require.NoError(t, err)

			gotResult, err := j.Target(tt.sourceInput, true)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
