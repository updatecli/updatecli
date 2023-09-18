package dockerimage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func TestCondition(t *testing.T) {
	tests := []struct {
		name           string
		spec           Spec
		source         string
		expectedError  bool
		expectedResult bool
	}{
		{
			name: "Success",
			spec: Spec{
				Image:         "ghcr.io/updatecli/updatecli",
				Architectures: []string{"amd64"},
			},
			source:         "v0.35.0",
			expectedResult: true,
		},
		{
			name: "Success - No architectures",
			spec: Spec{
				Image: "ghcr.io/updatecli/updatecli",
			},
			source:         "v0.35.0",
			expectedResult: true,
		},
		{
			name: "Success - operating system",
			spec: Spec{
				Image:           "eclipse-temurin",
				Architecture:    "amd64",
				OperatingSystem: "windows",
			},
			source:         "11.0.20.1_1-jdk-nanoserver-ltsc2022",
			expectedResult: true,
		},
		{
			name: "Failure - missing architecture",
			spec: Spec{
				Image:         "ghcr.io/updatecli/updatecli",
				Architectures: []string{"i386"},
			},
			source:         "v0.35.0",
			expectedResult: false,
		},
		{
			name: "Failure - one missing one not architecture",
			spec: Spec{
				Image:         "ghcr.io/updatecli/updatecli",
				Architectures: []string{"amd64", "i386"},
			},
			source:         "v0.35.0",
			expectedResult: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			require.NoError(t, err)

			gotResult := result.Condition{}
			err = got.Condition(tt.source, nil, &gotResult)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Pass)
		})
	}
}
