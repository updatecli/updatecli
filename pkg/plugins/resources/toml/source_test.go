package toml

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
			expectedErrorMsg: errors.New("cannot find value for path \".doNotExist\" from file \"testdata/data.toml\""),
		},
		{
			name: "Test array exist",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  ".database.ports.[1]",
			},
			expectedResult: "8001",
		},
		{
			name: "Test Query exist",
			spec: Spec{
				File:  "testdata/data.toml",
				Query: ".employees.[*].role",
				VersionFilter: version.Filter{
					Kind:    "regex",
					Pattern: "I(.*)",
				},
			},
			expectedResult: "IC",
		},
		{
			name: "Default successful workflow with Dasel v2",
			spec: Spec{
				File:   "testdata/data.toml",
				Key:    ".owner.firstName",
				Engine: strPtr(ENGINEDASEL_V2),
			},
			expectedResult: "Jack",
		},
		{
			name: "Array item with Dasel v2",
			spec: Spec{
				File:   "testdata/data.toml",
				Key:    ".database.ports.[1]",
				Engine: strPtr(ENGINEDASEL_V2),
			},
			expectedResult: "8001",
		},
		{
			name: "Default successful workflow with Dasel v3",
			spec: Spec{
				File:   "testdata/data.toml",
				Key:    "owner.firstName",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: "Jack",
		},
		{
			name: "Array item with Dasel v3",
			spec: Spec{
				File:   "testdata/data.toml",
				Key:    "database.ports[1]",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: "8001",
		},
		{
			name: "Nested array of objects with Dasel v3",
			spec: Spec{
				File:   "testdata/data.toml",
				Key:    "employees[3].role",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: "M",
		},
		{
			name: "Empty result with Dasel v3",
			spec: Spec{
				File:   "testdata/data.toml",
				Key:    "owner.surname",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			expectedResult: "",
		},
		{
			name: "Test key do not exist with Dasel v3",
			spec: Spec{
				File:   "testdata/data.toml",
				Key:    "doNotExist",
				Engine: strPtr(ENGINEDASEL_V3),
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("cannot find value for path \"doNotExist\" from file \"testdata/data.toml\""),
			expectedResult:   "",
		},
		{
			name: "Version filter with Dasel v3",
			spec: Spec{
				File:   "testdata/data.toml",
				Key:    "employees.map(role)...",
				Engine: strPtr(ENGINEDASEL_V3),
				VersionFilter: version.Filter{
					Kind:    "regex",
					Pattern: "I(.*)",
				},
			},
			expectedResult: "IC",
		},
		{
			name: "Bare dasel alias resolves to latest engine (v3)",
			spec: Spec{
				File:   "testdata/data.toml",
				Key:    "owner.firstName",
				Engine: strPtr(ENGINEDASEL_LATEST),
			},
			expectedResult: "Jack",
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {
			j, err := New(tt.spec)

			require.NoError(t, err)

			gotResult := result.Source{}
			err = j.Source(context.Background(), "", &gotResult)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Information)
		})
	}
}
