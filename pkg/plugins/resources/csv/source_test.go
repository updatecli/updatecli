package csv

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
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
				File:    "testdata/data.csv",
				Key:     ".[0].firstname",
				Comma:   ',',
				Comment: '#',
			},
			expectedResult: "John",
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
			expectedResult: "John",
		},
		{
			name: "Default successful workflow",
			spec: Spec{
				File:    "testdata/data.2.csv",
				Key:     ".[0].firstname",
				Comma:   ';',
				Comment: '#',
			},
			expectedResult: "John",
		},
		{
			name: "Default successful workflow with empty result",
			spec: Spec{
				File: "testdata/data.csv",
				Key:  ".[0].surname",
			},
			expectedResult: "",
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/data.csv",
				Key:   ".doNotExist",
				Value: "",
			},
			expectedResult:   "",
			wantErr:          true,
			expectedErrorMsg: errors.New("✗ cannot find value for path \".doNotExist\" from file \"testdata/data.csv\""),
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			c, err := New(tt.spec)

			require.NoError(t, err)

			gotResult, err := c.Source("")

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
