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
				Image:         "jenkinsciinfra/wiki",
				Architectures: []string{"amd64"},
			},
			source:         "0.1.6",
			expectedResult: true,
		},
		{
			name: "Failure - missing architecture",
			spec: Spec{
				Image:         "jenkinsciinfra/wiki",
				Architectures: []string{"arm64"},
			},
			source:         "0.1.6",
			expectedResult: false,
		},
		{
			name: "Failure - one missing one not architecture",
			spec: Spec{
				Image:         "jenkinsciinfra/wiki",
				Architectures: []string{"amd64", "arm64"},
			},
			source:         "0.1.6",
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
