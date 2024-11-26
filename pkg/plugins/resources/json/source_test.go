package json

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
				File: "testdata/data.json",
				Key:  ".firstName",
			},
			expectedResult: []result.SourceInformation{{
				Value: "Jack",
			}},
		},
		{
			name: "Default successful workflow with empty result",
			spec: Spec{
				File: "testdata/data.json",
				Key:  ".surname",
			},
			expectedResult: []result.SourceInformation{{
				Value: "",
			}},
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/data.json",
				Key:   ".doNotExist",
				Value: "",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("✗ cannot find value for path \".doNotExist\" from file \"testdata/data.json\""),
		},
		{
			name: "Test array exist",
			spec: Spec{
				File: "testdata/data.json",
				Key:  ".children.[1]",
			},
			expectedResult: []result.SourceInformation{{
				Value: "Thomas",
			}},
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
