package hcl

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/strings/slices"
)

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
				File: "testdata/data.hcl",
				Path: "resource.person.john.first_name",
			},
			workingDir:     "",
			expectedResult: []string{"testdata/data.hcl"},
		},
		{
			name: "Success - Working directory",
			spec: Spec{
				File: "testdata/data.hcl",
				Path: "resource.person.john.first_name",
			},
			workingDir:     "/tmp",
			expectedResult: []string{"/tmp/testdata/data.hcl"},
		},
		{
			name: "Success - Files with empty working directory",
			spec: Spec{
				Files: []string{"testdata/data.hcl", "testdata/data2.hcl"},
				Path:  "resource.person.john.first_name",
			},
			workingDir:     "",
			expectedResult: []string{"testdata/data.hcl", "testdata/data2.hcl"},
		},
		{
			name: "Success - Working directory",
			spec: Spec{
				Files: []string{"testdata/data.hcl", "testdata/data2.hcl"},
				Path:  "resource.person.john.first_name",
			},
			workingDir:     "/tmp",
			expectedResult: []string{"/tmp/testdata/data.hcl", "/tmp/testdata/data2.hcl"},
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
