package gomod

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestTarget(t *testing.T) {
	tests := []struct {
		name            string
		spec            Spec
		expectedChanged bool
		expectedError   bool
	}{
		{
			name: "Test module path exist and need no change",
			spec: Spec{
				File:    "testdata/go.mod",
				Module:  "sigs.k8s.io/yaml",
				Version: "v1.3.0",
			},
			expectedChanged: false,
		},
		{
			name: "Test module path exist and need change",
			spec: Spec{
				File:    "testdata/go.mod",
				Module:  "sigs.k8s.io/yaml",
				Version: "v2.0.0",
			},
			expectedChanged: true,
		},
		{
			name: "Test module path do not exist",
			spec: Spec{
				File:   "testdata/go.mod",
				Module: "doNotExist",
			},
			expectedChanged: false,
			expectedError:   true,
		},
		{
			name: "Ensure Go version should be updated",
			spec: Spec{
				File:    "testdata/go.mod",
				Version: "1.30",
			},
			expectedChanged: true,
		},
		{
			name: "Ensure Go version is already up to date",
			spec: Spec{
				File:    "testdata/go.mod",
				Version: "1.20",
			},
			expectedChanged: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			require.NoError(t, err)
			gotResult := result.Target{}

			err = got.Target("", nil, true, &gotResult)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedChanged, gotResult.Changed)
		})
	}

}
