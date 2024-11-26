package release

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
		wantResult []result.SourceInformation
		wantErr    bool
	}{
		{
			name: "repository updatecli/updatecli-nonexistent should not exist",
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
			wantErr: true,
		},
		{
			name: "repository updatecli/demo-terminal should exist but no release",
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
				Repository: "demo-terminal",
			},
			wantErr: true,
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
					Pattern: "~2.15",
				},
			},
			wantResult: []result.SourceInformation{{
				Key:   "",
				Value: "v2.15.0",
			}},
			wantErr: false,
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
