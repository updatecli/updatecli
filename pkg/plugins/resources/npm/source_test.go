package npm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestSource(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		spec           Spec
		expectedResult string
		expectedError  bool
	}{
		{
			name: "Passing case of retrieving axios versions ",
			spec: Spec{
				Name: "axios",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~0",
				},
			},
			expectedResult: "0.27.2",
			expectedError:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.spec)
			if tt.expectedError {
				assert.Error(t, err)
				return
			}
			gotVersion, err := got.Source("")
			require.NoError(t, err)
			assert.Equal(t, tt.expectedResult, gotVersion)
		})
	}

}
