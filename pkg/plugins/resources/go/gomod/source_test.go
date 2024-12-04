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
		expectedResult []result.SourceInformation
		expectedError  bool
	}{
		{
			name: "Test module path exist",
			spec: Spec{
				File:   "testdata/go.mod",
				Module: "sigs.k8s.io/yaml",
			},
			expectedResult: []result.SourceInformation{{
				Key:   "",
				Value: "v1.3.0",
			}},
		},
		{
			name: "Test module path do not exist",
			spec: Spec{
				File:   "testdata/go.mod",
				Module: "doNotExist",
			},
			expectedError: true,
		},
		{
			name: "Test retrieving indirect modulepath",
			spec: Spec{
				File:     "testdata/go.mod",
				Module:   "github.com/Azure/go-autorest/autorest/azure/auth",
				Indirect: true,
			},
			expectedResult: []result.SourceInformation{{
				Key:   "",
				Value: "v0.5.11",
			}},
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
			expectedResult: []result.SourceInformation{{
				Key:   "",
				Value: "v3.2.3",
			}},
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
