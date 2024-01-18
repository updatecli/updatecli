package gomodule

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
		expectedResult string
		expectedError  bool
	}{
		{
			spec: Spec{
				Module: "github.com/updatecli/updatecli",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "0.47",
				},
			},
			expectedResult: "v0.47.2",
		},
		{
			spec: Spec{
				Proxy:  "proxy.golang.org",
				Module: "github.com/updatecli/updatecli",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "0.47",
				},
			},
			expectedResult: "v0.47.2",
		},
		{
			spec: Spec{
				Proxy:  "direct,proxy.golang.org",
				Module: "github.com/updatecli/updatecli",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "0.47",
				},
			},
			expectedResult: "v0.47.2",
		},
		{
			spec: Spec{
				Module: "github.com/MakeNowJust/heredoc",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "1.0.0",
				},
			},
			expectedResult: "v1.0.0",
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
