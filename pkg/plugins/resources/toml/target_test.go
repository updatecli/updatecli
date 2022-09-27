package toml

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
			name: "Test key do not exist",
			spec: Spec{
				File:     "testdata/data.toml",
				Key:      ".doNotExist.[*]",
				Value:    "",
				Multiple: true,
			},
			expectedResult:   false,
			sourceInput:      "M",
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find multiple value for query \".doNotExist.[*]\" from file \"testdata/data.toml\""),
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/data.toml",
				Key:   ".doNotExist",
				Value: "",
			},
			expectedResult:   false,
			sourceInput:      "M",
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find value for query \".doNotExist\" from file \"testdata/data.toml\""),
		},
		{
			name: "Default successful multiple update workflow",
			spec: Spec{
				File:     "testdata/data.toml",
				Key:      ".employees.[*].role",
				Multiple: true,
			},
			sourceInput:    "M",
			expectedResult: true,
		},
		{
			name: "Successful conditional multiple update workflow",
			spec: Spec{
				File:     "testdata/data.toml",
				Key:      ".employees.(address=AU).role",
				Multiple: true,
			},
			sourceInput:    "M",
			expectedResult: false,
		},
		{
			name: "Successful multiple map update workflow",
			spec: Spec{
				File:     "testdata/data.toml",
				Key:      ".benefits.[0].country.(country=UK).name",
				Multiple: true,
			},
			sourceInput:    "all",
			expectedResult: true,
		},
		{
			name: "Successful single update workflow",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  ".owner.firstName",
			},
			sourceInput:    "Tom",
			expectedResult: true,
		},
		{
			name: "Successful no update workflow",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  ".owner.firstName",
			},
			sourceInput:    "Jack",
			expectedResult: false,
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
