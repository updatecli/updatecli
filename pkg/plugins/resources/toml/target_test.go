package toml

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
			name: "Deprecated multiple Test key do not exist",
			spec: Spec{
				File:     "testdata/data.toml",
				Key:      ".doNotExist.[*]",
				Value:    "",
				Multiple: true,
			},
			expectedResult:   false,
			sourceInput:      "M",
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find multiple value for query \".doNotExist.[*]\" from file \"testdata/data.toml\""),
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/data.toml",
				Query: ".doNotExist.[*]",
				Value: "",
			},
			expectedResult:   false,
			sourceInput:      "M",
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find multiple value for query \".doNotExist.[*]\" from file \"testdata/data.toml\""),
		},
		{
			name: "Test key do not exist",
			spec: Spec{
				File:  "testdata/data.toml",
				Key:   ".doNotExist",
				Value: "",
			},
			expectedResult:   false,
			sourceInput:      "M",
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find value for query \".doNotExist\" from file \"testdata/data.toml\""),
		},
		{
			name: "Default successful multiple update workflow",
			spec: Spec{
				File:  "testdata/data.toml",
				Query: ".employees.[*].role",
			},
			sourceInput:    "M",
			expectedResult: true,
		},
		{
			name: "Successful conditional multiple update workflow",
			spec: Spec{
				File:  "testdata/data.toml",
				Query: ".employees.(address=AU).role",
			},
			sourceInput:    "M",
			expectedResult: false,
		},
		{
			name: "Successful multiple map update workflow",
			spec: Spec{
				File:  "testdata/data.toml",
				Query: ".benefits.[0].country.(country=UK).name",
			},
			sourceInput:    "all",
			expectedResult: true,
		},
		{
			name: "Successful single update workflow",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  ".owner.firstName",
			},
			sourceInput:    "Tom",
			expectedResult: true,
		},
		{
			name: "Successful no update workflow",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  ".owner.firstName",
			},
			sourceInput:    "Jack",
			expectedResult: false,
		},
		{
			name: "Failing on non-existing key by default",
			spec: Spec{
				File: "testdata/data.toml",
				Key:  ".owner.age",
			},
			sourceInput:      "50",
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New("could not find value for query \".owner.age\" from file \"testdata/data.toml\""),
		},
		{
			name: "Successful update on non-existing key",
			spec: Spec{
				File:             "testdata/data.toml",
				Key:              ".owner.age",
				CreateMissingKey: true,
			},
			sourceInput:    "50",
			expectedResult: true,
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
