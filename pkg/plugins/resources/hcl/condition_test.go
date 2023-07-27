package hcl

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestCondition(t *testing.T) {
	testData := []struct {
		name             string
		spec             Spec
		source           string
		expectedResult   bool
		expectedErrorMsg error
		wantErr          bool
	}{
		{
			name: "Success - Using Value",
			spec: Spec{
				File:  "testdata/data.hcl",
				Key:   "resource.person.john.first_name",
				Value: "John",
			},
			source:         "",
			expectedResult: true,
		},
		{
			name: "Success - Using Source",
			spec: Spec{
				File: "testdata/data.hcl",
				Key:  "resource.person.john.first_name",
			},
			source:         "John",
			expectedResult: true,
		},
		{
			name: "Failure - Using Source",
			spec: Spec{
				File: "testdata/data.hcl",
				Key:  "resource.person.john.first_name",
			},
			source:         "Jack",
			expectedResult: false,
		},
		{
			name: "Success - Empty",
			spec: Spec{
				File: "testdata/data.hcl",
				Key:  "resource.person.john.middle_name",
			},
			source:         "",
			expectedResult: true,
		},
		{
			name: "Failure - File does not exists",
			spec: Spec{
				File: "testdata/doNotExist.hcl",
				Key:  "resource.person.john.first_name",
			},
			source:           "",
			wantErr:          true,
			expectedErrorMsg: errors.New("reading hcl file: ✗ The specified file \"testdata/doNotExist.hcl\" does not exist"),
		},
		{
			name: "Failure - Path does not exists",
			spec: Spec{
				File: "testdata/data.hcl",
				Key:  "resource.person.john.not_exist",
			},
			source:           "",
			wantErr:          true,
			expectedErrorMsg: errors.New("✗ cannot find value for path \"resource.person.john.not_exist\" from file \"testdata/data.hcl\""),
		},
		{
			name: "Failure - Multiple Files",
			spec: Spec{
				Files: []string{"testdata/data.hcl", "testdata/data2.hcl"},
				Key:   "resource.person.john.first_name",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("✗ HCL condition only supports one file"),
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			h, err := New(tt.spec)

			require.NoError(t, err)

			gotResult := result.Condition{}
			err = h.Condition(tt.source, nil, &gotResult)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Pass)
		})
	}
}
