package branch

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
			name: "v1 should exist",
			spec: Spec{
				Spec: client.Spec{
					URL:   "try.gitea.io",
					Token: "",
				},
				Owner:      "olblak",
				Repository: "updatecli-test",
				VersionFilter: version.Filter{
					Pattern: "v1",
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
