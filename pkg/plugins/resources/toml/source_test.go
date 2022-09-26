package toml

import (
	"errors"
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
			name: "Default successful workflow",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  ".owner.firstName",
			},
			expectedResult: "Jack",
		},
		{
			name: "Default successful workflow with empty result",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  ".owner.surname",
			},
			expectedResult: "",
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/data.toml",
				Key:   ".doNotExist",
				Value: "",
			},
			expectedResult:   "",
			wantErr:          true,
			expectedErrorMsg: errors.New("✗ could not find value for query \".doNotExist\" from file \"testdata/data.toml\""),
		},
		{
			name: "Test array exist",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  ".database.ports.[1]",
			},
			expectedResult: "8001",
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
