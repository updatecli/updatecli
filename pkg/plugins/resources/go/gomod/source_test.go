package gomod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSource(t *testing.T) {
	tests := []struct {
		name           string
		spec           Spec
		expectedResult string
		expectedError  bool
	}{
		{
			name: "Test module path exist",
			spec: Spec{
				File:   "testdata/go.mod",
				Module: "sigs.k8s.io/yaml",
			},
			expectedResult: "v1.3.0",
		},
		{
			name: "Test module path do not exist",
			spec: Spec{
				File:   "testdata/go.mod",
				Module: "doNotExist",
			},
			expectedResult: "",
			expectedError:  true,
		},
		{
			name: "Test retrieving indirect modulepath",
			spec: Spec{
				File:     "testdata/go.mod",
				Module:   "github.com/Azure/go-autorest/autorest/azure/auth",
				Indirect: true,
			},
			expectedResult: "v0.5.11",
		},
		{
			name: "Test modulepath not found because it's an indirect dep",
			spec: Spec{
				File:   "testdata/go.mod",
				Module: "github.com/Azure/go-autorest/autorest/azure/auth",
			},
			expectedError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			require.NoError(t, err)
			gotVersion, err := got.Source("")
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, gotVersion)
		})
	}

}
