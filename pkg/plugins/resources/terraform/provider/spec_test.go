package provider

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
				File:     "testdata/terraform.lock.hcl",
				Provider: "kubernetes",
			},
			wantErr: false,
		},
		{
			name: "Success - Files",
			spec: Spec{
				Files:    []string{"testdata/terraform.lock.hcl"},
				Provider: "kubernetes",
			},
			wantErr: false,
		},
		{
			name: "Failure - No file or files",
			spec: Spec{
				Provider: "kubernetes",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Failure - Both file and files",
			spec: Spec{
				File:     "testdata/terraform.lock.hcl",
				Files:    []string{"testdata/terraform.lock.hcl"},
				Provider: "kubernetes",
			},
			wantErr:          true,
			expectedErrorMsg: errors.New("wrong spec content"),
		},
		{
			name: "Failure - No provider",
			spec: Spec{
				File: "testdata/terraform.lock.hcl",
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
