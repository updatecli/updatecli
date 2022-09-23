package json

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
)

func TestSource(t *testing.T) {

	testData := []struct {
		name             string
		spec             Spec
		expectedResult   string
		expectedErrorMsg error
		wantErr          bool
	}{
		{
			name: "Default successfull workflow",
			spec: Spec{
				File: "testdata/data.json",
				Key:  ".firstName",
			},
			expectedResult: "Jack",
		},
		{
			name: "Default successful workflow with empty result",
			spec: Spec{
				File: "testdata/data.json",
				Key:  ".surname",
			},
			expectedResult: "",
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/data.json",
				Key:   ".doNotExist",
				Value: "",
			},
			expectedResult: "",
		},
		{
			name: "Test array exist",
			spec: Spec{
				File: "testdata/data.json",
				Key:  ".children.[1]",
			},
			expectedResult: "Thomas",
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			j, err := New(tt.spec)

			require.NoError(t, err)

			gotResult, err := j.Source("")

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
