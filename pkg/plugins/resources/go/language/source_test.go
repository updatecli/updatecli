package language

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			spec: Spec{
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "1.13.3",
				},
			},
			expectedResult: "1.13.3",
		},
		{
			spec: Spec{
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "1.17.0",
				},
			},
			expectedResult: "1.17.0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec, false)
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
