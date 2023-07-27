package hcl

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
			name: "Success - No change",
			spec: Spec{
				File: "testdata/data.hcl",
				Key:  "resource.person.john.first_name",
			},
			sourceInput:    "John",
			expectedResult: false,
		},
		{
			name: "Success - Expected change",
			spec: Spec{
				File: "testdata/data.hcl",
				Key:  "resource.person.john.first_name",
			},
			sourceInput:    "Jonathan",
			expectedResult: true,
		},
		{
			name: "Success - Expected change using Value",
			spec: Spec{
				File:  "testdata/data.hcl",
				Key:   "resource.person.john.first_name",
				Value: "Jonathan",
			},
			expectedResult: true,
		},
		{
			name: "Success - Expected change replacing empty value",
			spec: Spec{
				File: "testdata/data.hcl",
				Key:  "resource.person.john.middle_name",
			},
			sourceInput:    "Joe",
			expectedResult: true,
		},
		{
			name: "Failure - File does not exists",
			spec: Spec{
				File: "testdata/doNotExist.hcl",
				Key:  "resource.person.john.first_name",
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New("reading hcl file: ✗ The specified file \"testdata/doNotExist.hcl\" does not exist"),
		},
		{
			name: "Failure - Path does not exists",
			spec: Spec{
				File: "testdata/data.hcl",
				Key:  "resource.person.john.not_exist",
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New("✗ cannot find value for path \"resource.person.john.not_exist\" from file \"testdata/data.hcl\""),
		},
		{
			name: "Failure - HTTP Target",
			spec: Spec{
				File: "http://localhost/doNotExist.hcl",
				Key:  "resource.person.john.first_name",
			},
			expectedResult:   false,
			wantErr:          true,
			expectedErrorMsg: errors.New("✗ URL scheme is not supported for HCL target: \"http://localhost/doNotExist.hcl\""),
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
