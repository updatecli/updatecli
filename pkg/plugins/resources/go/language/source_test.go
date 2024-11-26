package language

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
		expectedResult []result.SourceInformation
		expectedError  bool
	}{
		{
			spec: Spec{
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "1.13.3",
				},
			},
			expectedResult: []result.SourceInformation{{
				Key:   "",
				Value: "1.13.3",
			}},
		},
		{
			spec: Spec{
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "1.17.0",
				},
			},
			expectedResult: []result.SourceInformation{{
				Key:   "",
				Value: "1.17.0",
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
