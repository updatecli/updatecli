package csv

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func strPtr(s string) *string {
	return &s
}

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
			expectedErrorMsg: errors.New("cannot find value for path \".doNotExist\" from file \"testdata/data.csv\""),
		},
		{
			name: "Default successful workflow with Dasel v2",
			spec: Spec{
				File:   "testdata/data.csv",
				Key:    ".[0].firstname",
				Engine: strPtr(ENGINEDASEL_V2),
			},
			expectedResult: "John",
		},
		{
			name: "Default successful workflow with Dasel v3",
			spec: Spec{
				File:   "testdata/data.csv",
				Key:    "$this[0].firstname",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: "John",
		},
		{
			name: "Row of a different column with Dasel v3",
			spec: Spec{
				File:   "testdata/data.csv",
				Key:    "$this[2].surname",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: "Olblak",
		},
		{
			name: "Empty result with Dasel v3",
			spec: Spec{
				File:   "testdata/data.csv",
				Key:    "$this[0].surname",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: "",
		},
		{
			name: "Test key do not exist with Dasel v3",
			spec: Spec{
				File:   "testdata/data.csv",
				Key:    "$this[0].doNotExist",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult:   "",
			wantErr:          true,
			expectedErrorMsg: errors.New("cannot find value for path \"$this[0].doNotExist\" from file \"testdata/data.csv\""),
		},
		{
			name: "Version filter across rows with Dasel v3",
			spec: Spec{
				File:   "testdata/data.csv",
				Key:    "map(firstname)...",
				Engine: strPtr(ENGINEDASEL_V3),
				VersionFilter: version.Filter{
					Kind:    "regex",
					Pattern: "^Jo",
				},
			},
			expectedResult: "John",
		},
		{
			name: "Bare dasel alias resolves to latest engine (v3)",
			spec: Spec{
				File:   "testdata/data.csv",
				Key:    "$this[0].firstname",
				Engine: strPtr(ENGINEDASEL_LATEST),
			},
			expectedResult: "John",
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			c, err := New(tt.spec)

			require.NoError(t, err)

			gotResult := result.Source{}
			err = c.Source(context.Background(), "", &gotResult)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Information)
		})
	}
}
