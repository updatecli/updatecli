package gomodule

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeGoModuleName(t *testing.T) {

	tests := []struct {
		name           string
		module         string
		expectedResult string
	}{
		{
			name:           "check that uppercase character are correctly lowercase with !",
			module:         "github.com/UpdateCli/updateclI",
			expectedResult: "github.com/!update!cli/updatecl!i",
		},
		{
			name:           "check that uppercase character are correctly lowercase with !",
			module:         "github.com/updatecli/updatecli",
			expectedResult: "github.com/updatecli/updatecli",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult := sanitizeGoModuleNameForProxy(tt.module)
			assert.Equal(t, tt.expectedResult, gotResult)
		})
	}
}
