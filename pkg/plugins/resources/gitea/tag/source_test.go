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
			name: "repository olblak/updatecli should not exist",
			manifest: struct {
				URL           string
				Token         string
				Owner         string
				Repository    string
				VersionFilter version.Filter
			}{
				URL:        "codeberg.org",
				Token:      "",
				Owner:      "updatecli",
				Repository: "updatecli-nonexistent",
			},
			wantResult: "",
			wantErr:    true,
		},
		{
			name: "repository should exist with release 0.0.3",
			manifest: struct {
				URL           string
				Token         string
				Owner         string
				Repository    string
				VersionFilter version.Filter
			}{
				URL:        "codeberg.org",
				Token:      "",
				Owner:      "updatecli",
				Repository: "updatecli-action",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~1",
				},
			},
			wantResult: "v1.33.0",
			wantErr:    false,
		},
		{
			name: "repository should exist with no release 1.0.0",
			manifest: struct {
				URL           string
				Token         string
				Owner         string
				Repository    string
				VersionFilter version.Filter
			}{
				URL:        "codeberg.org",
				Token:      "",
				Owner:      "updatecli",
				Repository: "updatecli-action",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "~0",
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
