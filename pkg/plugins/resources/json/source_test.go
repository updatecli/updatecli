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
		isList           bool
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
		{
			name: "Retrieve json source as list with query",
			spec: Spec{
				File:  "testdata/data.json",
				Query: ".children.[*]",
			},
			isList: true,
			expectedResult: []result.SourceInformation{{
				Key:   "0",
				Value: "Catherine",
			}, {
				Key:   "1",
				Value: "Thomas",
			}, {
				Key:   "2",
				Value: "Trevor",
			}},
		},
		{
			name: "Retrieve json source as list with key should error",
			spec: Spec{
				File: "testdata/data.json",
				Key:  ".children.[*]",
			},
			isList:           true,
			wantErr:          true,
			expectedErrorMsg: errors.New("✗ json source can only be configured as a list with the `query` param."),
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			j, err := New(tt.spec, tt.isList)

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
