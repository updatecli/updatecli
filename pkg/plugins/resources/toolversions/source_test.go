package toolversions

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestSource(t *testing.T) {

	testData := []struct {
		name             string
		spec             Spec
		expectedResult   []result.SourceInformation
		expectedErrorMsg error
		wantErr          bool
	}{
		{
			name: "Default successful workflow",
			spec: Spec{
				File: "testdata/.tool-versions",
				Key:  "bats",
			},
			expectedResult: []result.SourceInformation{{
				Value: "1.0.0",
			}},
		},
		{
			name: "Default successful workflow with empty result",
			spec: Spec{
				File: "testdata/.tool-versions",
				Key:  "empty",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find value for key \"empty\" from file \"testdata/.tool-versions\""),
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/.tool-versions",
				Key:   "doNotExist",
				Value: "",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find value for key \"doNotExist\" from file \"testdata/.tool-versions\""),
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			j, err := New(tt.spec)

			require.NoError(t, err)

			gotResult := result.Source{}
			err = j.Source("", &gotResult)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Information)
		})
	}
}
