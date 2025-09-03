package gomod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
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
		{
			name: "Test retrieving module from https",
			spec: Spec{
				File:   "https://raw.githubusercontent.com/updatecli/updatecli/v0.60.0/go.mod",
				Module: "github.com/Masterminds/sprig/v3",
			},
			expectedResult: "v3.2.3",
		},
		{
			name: "Test retrieving module from replace",
			spec: Spec{
				File:    "testdata/replace.go.mod",
				Module:  "github.com/gin-gonic/gin",
				Replace: true,
			},
			expectedResult: "v1.7.0",
		},
		{
			name: "Test version downgrade",
			spec: Spec{
				File:    "testdata/replace.2.go.mod",
				Module:  "github.com/crewjam/saml",
				Replace: true,
			},
			expectedResult: "v0.5.0",
		},
		{
			name: "Test module replacement",
			spec: Spec{
				File:    "testdata/replace.2.go.mod",
				Module:  "github.com/rancher/saml",
				Replace: true,
			},
			expectedResult: "v0.2.0",
		},
		{
			name: "Test dev module",
			spec: Spec{
				File:    "testdata/replace.2.go.mod",
				Module:  "github.com/stretchr/testify",
				Replace: true,
			},
			expectedError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			require.NoError(t, err)
			gotResult := result.Source{}
			err = got.Source("", &gotResult)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, gotResult.Information)
		})
	}

}
