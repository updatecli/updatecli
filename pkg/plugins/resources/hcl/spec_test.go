package hcl

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	testData := []struct {
		name             string
		spec             Spec
		expectedErrorMsg error
		wantErr          bool
	}{
		{
			name: "Success - File",
			spec: Spec{
				File: "testdata/data.hcl",
				Path: "resource.person.john.first_name",
			},
			wantErr: false,
		},
		{
			name: "Success - Files",
			spec: Spec{
				Files: []string{"testdata/data.hcl"},
				Path:  "resource.person.john.first_name",
			},
			wantErr: false,
		},
		{
			name: "Failure - No file or files",
			spec: Spec{
				Path: "resource.person.john.first_name",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Failure - No path",
			spec: Spec{
				File: "testdata/data.hcl",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Failure - Both file and files",
			spec: Spec{
				File:  "testdata/data.hcl",
				Files: []string{"testdata/data.hcl"},
				Path:  "resource.person.john.first_name",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
