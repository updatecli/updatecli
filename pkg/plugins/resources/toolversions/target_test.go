package toolversions

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
			name: "Successful update workflow with a new key",
			spec: Spec{
				File:             "testdata/.tool-versions",
				Key:              "doNotExist",
				Value:            "1.0.0",
				CreateMissingKey: true,
			},
			expectedResult: true,
			sourceInput:    "1.0.0",
			wantErr:        false,
		},
		{
			name: "Successful update workflow with an existing key and different value",
			spec: Spec{
				File: "testdata/.tool-versions",
				Key:  "bats",
			},
			sourceInput:    "2.0.0",
			expectedResult: true,
		},
		{
			name: "Successful no update workflow",
			spec: Spec{
				File: "testdata/.tool-versions",
				Key:  "golang",
			},
			expectedResult: false,
			sourceInput:    "1.8.2",
		},
		{
			name: "Test file do not exist",
			spec: Spec{
				File: ".new",
				Key:  "golang",
			},
			expectedResult:   false,
			sourceInput:      "M",
			wantErr:          true,
			expectedErrorMsg: errors.New("file \".new\" does not exist"),
		},
		{
			name: "Failing on non-existing key by default",
			spec: Spec{
				File: "testdata/.tool-versions",
				Key:  "doNotExist",
			},
			expectedResult:   false,
			sourceInput:      "M",
			wantErr:          true,
			expectedErrorMsg: errors.New("key \"doNotExist\" does not exist. Use createMissingKey if you want to create the key"),
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
