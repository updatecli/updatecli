package tag

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func TestCondition(t *testing.T) {

	tests := []struct {
		name       string
		spec       Spec
		wantResult bool
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
			wantResult: false,
			wantErr:    true,
		},
		{
			name: "repository olblak/updatecli-mirror should exist with tags",
			spec: Spec{
				Spec: client.Spec{
					URL:   "try.gitea.io",
					Token: "",
				},
				Owner:      "olblak",
				Repository: "updatecli-mirror",
			},
			wantResult: true,
			wantErr:    false,
		},
		{
			name: "repository should exist with release 0.0.1",
			spec: Spec{
				Spec: client.Spec{
					URL:   "try.gitea.io",
					Token: "",
				},
				Owner:      "olblak",
				Repository: "updatecli-test",
			},
			wantResult: true,
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
			wantResult: false,
			wantErr:    true,
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
					Pattern: "0.0.1",
				},
			},
			wantResult: true,
			wantErr:    false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			g, _ := New(tt.spec)
			gotResult, gotErr := g.Condition("")

			if tt.wantErr {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}

			assert.Equal(t, tt.wantResult, gotResult)

		})

	}
}
