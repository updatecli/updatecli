package dockerimage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestSource(t *testing.T) {
	tests := []struct {
		name           string
		spec           Spec
		expectedError  bool
		expectedResult string
	}{
		{
			name: "Success",
			spec: Spec{
				Image: "jenkinsciinfra/wiki",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "0.1.6",
				},
				Architectures: []string{"amd64"},
			},
			expectedResult: "0.1.6",
			expectedError:  false,
		},
		{
			name: "Failure - missing architecture",
			spec: Spec{
				Image: "jenkinsciinfra/wiki",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "0.1.6",
				},
				Architectures: []string{"arm64"},
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
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedResult, gotResult.Information)
		})
	}
}
