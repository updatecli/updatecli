package csv

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
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
				File:    "testdata/data.csv",
				Key:     ".[0].firstname",
				Comma:   ',',
				Comment: '#',
			},
			expectedResult: []result.SourceInformation{{
				Value: "John",
			}},
		},
		{
			name: "Regex versionFilter successful workflow",
			spec: Spec{
				File:  "testdata/data.csv",
				Query: ".[*].firstname",
				VersionFilter: version.Filter{
					Kind:    "regex",
					Pattern: "^Jo",
				},
			},
			expectedResult: []result.SourceInformation{{
				Value: "John",
			}},
		},
		{
			name: "Default successful workflow",
			spec: Spec{
				File:    "testdata/data.2.csv",
				Key:     ".[0].firstname",
				Comma:   ';',
				Comment: '#',
			},
			expectedResult: []result.SourceInformation{{
				Key:   "",
				Value: "John",
			}},
		},
		{
			name: "Default successful workflow with empty result",
			spec: Spec{
				File: "testdata/data.csv",
				Key:  ".[0].surname",
			},
			expectedResult: []result.SourceInformation{{Key: "", Value: ""}},
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/data.csv",
				Key:   ".doNotExist",
				Value: "",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("cannot find value for path \".doNotExist\" from file \"testdata/data.csv\""),
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			c, err := New(tt.spec)

			require.NoError(t, err)

			gotResult := result.Source{}
			err = c.Source("", &gotResult)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Information)
		})
	}
}
