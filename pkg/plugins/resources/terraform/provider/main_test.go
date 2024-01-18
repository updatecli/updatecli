package provider

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuery(t *testing.T) {
	testData := []struct {
		name             string
		spec             Spec
		workingDir       string
		expectedErrorMsg error
		wantErr          bool
		expectedResult   string
	}{
		{
			name: "Success - Query file",
			spec: Spec{
				File:     "testdata/versions.tf",
				Provider: "azurerm",
			},
			wantErr:        false,
			workingDir:     "",
			expectedResult: `3.69.0`,
		},
		{
			name: "Success - Query files",
			spec: Spec{
				Files:    []string{"testdata/versions.tf"},
				Provider: "azurerm",
			},
			wantErr:        false,
			workingDir:     "",
			expectedResult: `3.69.0`,
		},
		{
			name: "Failure - missing",
			spec: Spec{
				File:     "testdata/versions.tf",
				Provider: "null",
			},
			wantErr:          true,
			workingDir:       "",
			expectedErrorMsg: errors.New(`✗ cannot find value for "null" from file "testdata/versions.tf"`),
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			h, err := New(tt.spec)

			require.NoError(t, err)

			err = h.Read()

			require.NoError(t, err)

			resourceFile := h.files["testdata/versions.tf"]

			resultVersion, err := h.Query(resourceFile)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, resultVersion)
		})
	}
}

func TestApply(t *testing.T) {
	testData := []struct {
		name             string
		spec             Spec
		value            string
		workingDir       string
		expectedErrorMsg error
		wantErr          bool
		expectedResult   string
	}{
		{
			name: "Success - Update",
			spec: Spec{
				File:     "testdata/versions.tf",
				Provider: "azurerm",
			},
			value:      "3.70.0",
			wantErr:    false,
			workingDir: "",
			expectedResult: `terraform {
  required_version = "~> 1.5"
  required_providers {
    azuread = {
      source  = "hashicorp/azuread"
      version = "2.41.0"
    }
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "3.70.0"
    }
  }
}
`,
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			h, err := New(tt.spec)

			require.NoError(t, err)

			err = h.Read()

			require.NoError(t, err)

			err = h.Apply(tt.spec.File, tt.value)

			require.NoError(t, err)

			if tt.wantErr {
				assert.Equal(t, tt.expectedErrorMsg.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, h.files[tt.spec.File].content)
		})
	}
}

func TestUpdateAbsoluteFilePath(t *testing.T) {
	testData := []struct {
		name           string
		spec           Spec
		workingDir     string
		expectedResult []string
	}{
		{
			name: "Success - Empty working directory",
			spec: Spec{
				File:     "testdata/versions.tf",
				Provider: "azurerm",
			},
			workingDir:     "",
			expectedResult: []string{"testdata/versions.tf"},
		},
		{
			name: "Success - Working directory",
			spec: Spec{
				File:     "testdata/versions.tf",
				Provider: "azurerm",
			},
			workingDir:     "/tmp",
			expectedResult: []string{"/tmp/testdata/versions.tf"},
		},
		{
			name: "Success - Files with empty working directory",
			spec: Spec{
				Files:    []string{"testdata/versions.tf", "testdata/data2.hcl"},
				Provider: "azurerm",
			},
			workingDir:     "",
			expectedResult: []string{"testdata/versions.tf", "testdata/data2.hcl"},
		},
		{
			name: "Success - Working directory",
			spec: Spec{
				Files:    []string{"testdata/versions.tf", "testdata/data2.hcl"},
				Provider: "azurerm",
			},
			workingDir:     "/tmp",
			expectedResult: []string{"/tmp/testdata/versions.tf", "/tmp/testdata/data2.hcl"},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			h, err := New(tt.spec)

			require.NoError(t, err)

			h.UpdateAbsoluteFilePath(tt.workingDir)

			for _, v := range h.files {
				assert.True(t, slices.Contains(tt.expectedResult, v.filePath), fmt.Sprintf("%s not in %v", v.filePath, tt.expectedResult))
			}
		})
	}
}
