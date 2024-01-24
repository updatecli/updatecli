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
				Image: "ghcr.io/updatecli/updatecli",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "v0.35.0",
				},
				Architectures: []string{"amd64"},
			},
			expectedResult: "v0.35.0",
			expectedError:  false,
		},
		{
			name: "Failure no version found",
			spec: Spec{
				Image: "ghcr.io/updatecli/updatecli",
				VersionFilter: version.Filter{
					Kind: "latest",
				},
				TagFilter:     "donotExist*",
				Architectures: []string{"amd64"},
			},
			expectedError: true,
		},
		{
			name: "Success - no architecture",
			spec: Spec{
				Image: "ghcr.io/updatecli/updatecli",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "v0.35.0",
				},
			},
			expectedResult: "v0.35.0",
			expectedError:  false,
		},
		{
			name: "Failure - missing architecture",
			spec: Spec{
				Image: "ghcr.io/updatecli/updatecli",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "v0.35.0",
				},
				Architectures: []string{"i386"},
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
