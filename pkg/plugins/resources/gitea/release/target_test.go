package release

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
)

func TestTarget(t *testing.T) {

	tests := []struct {
		name       string
		spec       Spec
		wantResult bool
		wantErr    bool
	}{
		// No token provided should error
		{
			name: "repository should exist with no release 1.0.0",
			spec: Spec{
				Spec: client.Spec{
					URL:   "try.gitea.io",
					Token: "tokenRequired",
				},
				Owner:      "olblak",
				Repository: "updatecli-test",
				Title:      "0.0.2",
				Tag:        "0.0.2",
			},
			// It's difficult to automatically test release creation without mocktesting
			wantResult: false,
			wantErr:    false,
		},
		{
			name: "repository should exist with no release 1.0.0",
			spec: Spec{
				Spec: client.Spec{
					URL:   "try.gitea.io",
					Token: "tokenRequired",
				},
				Owner:      "olblak",
				Repository: "updatecli-test",
				Title:      "0.0.3",
				Tag:        "0.0.3",
			},
			// It's difficult to automatically test release creation without mocktesting
			wantResult: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			g, _ := New(tt.spec)
			gotResult, gotErr := g.Target("", false)

			if tt.wantErr {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}

			assert.Equal(t, tt.wantResult, gotResult)

		})

	}
}
