package hcl

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
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
			name: "Success - Query",
			spec: Spec{
				File: "testdata/data.hcl",
				Path: "resource.person.john.first_name",
			},
			expectedResult: []result.SourceInformation{{
				Value: "John",
			}},
		},
		{
			name: "Success - Query empty value",
			spec: Spec{
				File: "testdata/data.hcl",
				Path: "resource.person.john.middle_name",
			},
			expectedResult: []result.SourceInformation{{
				Value: "",
			}},
		},
		{
			name: "Success - Query integer",
			spec: Spec{
				File: "testdata/data.hcl",
				Path: "resource.person.john.age",
			},
			expectedResult: []result.SourceInformation{{
				Value: "30",
			}},
		},
		{
			name: "Success - Query nested",
			spec: Spec{
				File: "testdata/data.hcl",
				Path: "resource.person.john.nested.attr1",
			},
			expectedResult: []result.SourceInformation{{
				Value: "val1",
			}},
		},
		{
			name: "Success - Query using Files",
			spec: Spec{
				Files: []string{"testdata/data.hcl"},
				Path:  "resource.person.john.first_name",
			},
			expectedResult: []result.SourceInformation{{
				Value: "John",
			}},
		},
		{
			name: "Failure - File does not exists",
			spec: Spec{
				File: "testdata/doNotExist.hcl",
				Path: "resource.person.john.first_name",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("reading hcl file: ✗ The specified file \"testdata/doNotExist.hcl\" does not exist"),
		},
		{
			name: "Failure - Path does not exists",
			spec: Spec{
				File: "testdata/data.hcl",
				Path: "resource.person.john.not_exist",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("✗ cannot find value for path \"resource.person.john.not_exist\" from file \"testdata/data.hcl\""),
		},
		{
			name: "Failure - Multiple Files",
			spec: Spec{
				Files: []string{"testdata/data.hcl", "testdata/data2.hcl"},
				Path:  "resource.person.john.first_name",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("✗ HCL source only supports one file"),
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			h, err := New(tt.spec)

			require.NoError(t, err)

			gotResult := result.Source{}
			err = h.Source("", &gotResult)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Information)
		})
	}
}
