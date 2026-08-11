package language

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestSource(t *testing.T) {
	tests := []struct {
		name           string
		spec           Spec
		expectedResult string
		expectedError  bool
	}{
		{
			name: "TestSourceWithVersionFilter",
			spec: Spec{
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "1.13.3",
				},
			},
			expectedResult: "1.13.3",
		},
		{
			name: "TestSourceWithVersionFilter2",
			spec: Spec{
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "1.17.0",
				},
			},
			expectedResult: "1.17.0",
		},
		{
			name: "TestSourceWithVersionFilterAndMinimumAge",
			spec: Spec{
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "1.19",
				},
				Age: age.Spec{
					Minimum: "1y",
				},
			},
			expectedResult: "1.19.13",
		},
		{
			name: "TestSourceWithVersionFilterAndMaximumAge",
			spec: Spec{
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "1.19",
				},
				Age: age.Spec{
					Maximum: "100y",
				},
			},
			expectedResult: "1.19.13",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			require.NoError(t, err)
			gotResult := result.Source{}
			err = got.Source(context.Background(), "", &gotResult)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, gotResult.Information)
		})
	}

}
