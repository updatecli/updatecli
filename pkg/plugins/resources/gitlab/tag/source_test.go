package tag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestSource(t *testing.T) {

	tests := []struct {
		name     string
		manifest struct {
			URL           string
			Token         string
			Owner         string
			Repository    string
			VersionFilter version.Filter
		}
		wantResult string
		wantErr    bool
	}{
		{
			name: "repository olblak/updatecli-nonexistent should not exist",
			manifest: struct {
				URL           string
				Token         string
				Owner         string
				Repository    string
				VersionFilter version.Filter
			}{
				URL:        "gitlab.com",
				Token:      "",
				Owner:      "olblak",
				Repository: "updatecli-nonexistent",
			},
			wantResult: "",
			wantErr:    true,
		},
		{
			name: "repository should exist with tag 0.3.0 without specifying gitlab.com",
			manifest: struct {
				URL           string
				Token         string
				Owner         string
				Repository    string
				VersionFilter version.Filter
			}{
				Token:      "",
				Owner:      "olblak",
				Repository: "updatecli",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "0.3.0",
				},
			},
			wantResult: "v0.3.0",
			wantErr:    false,
		},
		{
			name: "repository should exist with tag 0.3.0",
			manifest: struct {
				URL           string
				Token         string
				Owner         string
				Repository    string
				VersionFilter version.Filter
			}{
				URL:        "gitlab.com",
				Token:      "",
				Owner:      "olblak",
				Repository: "updatecli",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "0.3.0",
				},
			},
			wantResult: "v0.3.0",
			wantErr:    false,
		},
		{
			name: "repository should exist with no tag 0.3.11",
			manifest: struct {
				URL           string
				Token         string
				Owner         string
				Repository    string
				VersionFilter version.Filter
			}{
				URL:        "gitlab.com",
				Token:      "",
				Owner:      "olblak",
				Repository: "updatecli",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "v0.3.11",
				},
			},
			wantResult: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			g, gotErr := New(tt.manifest)
			require.NoError(t, gotErr)

			gotResult := result.Source{}
			gotErr = g.Source("", &gotResult)

			if tt.wantErr {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}

			assert.Equal(t, tt.wantResult, gotResult.Information)

		})

	}
}
