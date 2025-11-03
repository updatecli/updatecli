package json

import (
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
			name: "Get last array item successful workflow",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    ".phoneNumbers.all().filter(equal(type,office)).number",
				Engine: strPtr(ENGINEDASEL_V2),
			},
			expectedResult: "646 555-4567",
		},
		{
			name: "Default successful workflow",
			spec: Spec{
				File: "testdata/data.json",
				Key:  ".firstName",
			},
			expectedResult: "Jack",
		},
		{
			name: "Default successful workflow",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    ".firstName",
				Engine: strPtr(ENGINEDASEL_V2),
			},
			expectedResult: "Jack",
		},
		{
			name: "Get last array item successful workflow",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    ".phoneNumbers.last().type",
				Engine: strPtr(ENGINEDASEL_V2),
			},
			expectedResult: "office",
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
			name: "Default successful workflow with empty result",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    ".surname",
				Engine: strPtr(ENGINEDASEL_V2),
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
			wantErr:          true,
			expectedErrorMsg: errors.New("✗ cannot find value for path \".doNotExist\" from file \"testdata/data.json\""),
			expectedResult:   "",
		},
		{
			name: "Test key do not exist with Dasel v2",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    ".doNotExist",
				Value:  "",
				Engine: strPtr(ENGINEDASEL_V2),
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("✗ cannot find value for path \".doNotExist\" from file \"testdata/data.json\""),
			expectedResult:   "",
		},
		{
			name: "Test array exist",
			spec: Spec{
				File: "testdata/data.json",
				Key:  ".children.[1]",
			},
			expectedResult: "Thomas",
		},
		{
			name: "Test array exist with Dasel v2",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    ".children.[1]",
				Engine: strPtr(ENGINEDASEL_V2),
			},
			expectedResult: "Thomas",
		},
		{
			name: "Version filter with Dasel v2",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    ".children.all()",
				Engine: strPtr(ENGINEDASEL_V2),
				VersionFilter: version.Filter{
					Kind: "latest",
				},
			},
			expectedResult: "Trevor",
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
