package toolversions

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
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
				File:  "testdata/.tool-versions",
				Key:   "doNotExist",
				Value: "1.0.0",
			},
			expectedResult: true,
			sourceInput:    "M",
			wantErr:        false,
		},
		{
			name: "Successful update workflow",
			spec: Spec{
				File: "testdata/.tool-versions",
				Key:  "bats",
			},
			sourceInput:    "2.0.0",
			expectedResult: true,
		},
		{
			name: "Successful update workflow",
			spec: Spec{
				File: "testdata/.tool-versions",
				Key:  "commentedtool",
			},
			expectedResult:   false,
			sourceInput:      "M",
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find value for query \"commentedtool\" from file \"testdata/.tool-versions\""),
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			j, err := New(tt.spec)

			require.NoError(t, err)

			gotResult := result.Target{}

			err = j.Target(tt.sourceInput, nil, true, &gotResult)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Changed)
		})
	}
}
