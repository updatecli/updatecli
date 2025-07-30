package json

import (
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
			name: "Default successful workflow",
			spec: Spec{
				File: "testdata/data.json",
				Key:  ".firstName",
			},
			sourceInput:    "Jack",
			expectedResult: false,
		},
		{
			name: "Default successful workflow using Dasel v2",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    ".firstName",
				Engine: strPtr(ENGINEDASEL_V2),
			},
			sourceInput:    "Jack",
			expectedResult: false,
		},
		{
			name: "Default successful workflow",
			spec: Spec{
				File: "testdata/data.json",
				Key:  ".firstName",
			},
			sourceInput:    "Tom",
			expectedResult: true,
		},
		{
			name: "Default successful workflow",
			spec: Spec{
				File:  "testdata/data.json",
				Query: ".phoneNumbers.[*].type",
			},
			sourceInput:    "Tom",
			expectedResult: true,
		},
		{
			name: "Update first array item successful workflow",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    ".phoneNumbers.first().type",
				Engine: strPtr(ENGINEDASEL_V2),
			},
			sourceInput:    "apartment",
			expectedResult: true,
		},
		{
			name: "Unchanged first array item successful workflow",
			spec: Spec{
				File:   "testdata/data.json",
				Key:    ".phoneNumbers.first().type",
				Engine: strPtr(ENGINEDASEL_V2),
			},
			sourceInput:    "home",
			expectedResult: false,
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
