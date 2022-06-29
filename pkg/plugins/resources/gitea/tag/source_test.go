package tag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestSource(t *testing.T) {

	tests := []struct {
		name       string
		spec       Spec
		wantResult string
		wantErr    bool
	}{
		{
			name: "repository olblak/updatecli should not exist",
			spec: Spec{
				Spec: client.Spec{
					URL:   "try.gitea.io",
					Token: "",
				},
				Owner:      "olblak",
				Repository: "updatecli",
			},
			wantResult: "",
			wantErr:    true,
		},
		{
			name: "repository olblak/updatecli-mirror should exist but no release",
			spec: Spec{
				Spec: client.Spec{
					URL:   "try.gitea.io",
					Token: "",
				},
				Owner:      "olblak",
				Repository: "updatecli-mirror",
				VersionFilter: version.Filter{
					Kind: "semver",
				},
			},
			// For some reason, it sort the list of version alphabetically instead of by tag creation
			wantResult: "0.29.0",
			wantErr:    false,
		},
		{
			name: "repository should exist with release 0.0.3",
			spec: Spec{
				Spec: client.Spec{
					URL:   "try.gitea.io",
					Token: "",
				},
				Owner:      "olblak",
				Repository: "updatecli-test",
			},
			wantResult: "0.0.3",
			wantErr:    false,
		},
		{
			name: "repository should exist with no release 1.0.0",
			spec: Spec{
				Spec: client.Spec{
					URL:   "try.gitea.io",
					Token: "",
				},
				Owner:      "olblak",
				Repository: "updatecli-test",
				VersionFilter: version.Filter{
					Kind:    "semver",
					Pattern: "1.0.0",
				},
			},
			wantResult: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			g, _ := New(tt.spec)
			gotResult, gotErr := g.Source("")

			if tt.wantErr {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}

			assert.Equal(t, tt.wantResult, gotResult)

		})

	}
}
