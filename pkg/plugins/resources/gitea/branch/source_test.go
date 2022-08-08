package branch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSource(t *testing.T) {

	tests := []struct {
		name     string
		manifest struct {
			URL        string
			Token      string
			Owner      string
			Repository string
		}
		wantResult string
		wantErr    bool
	}{
		{
			name: "repository olblak/updatecli should not exist",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
			}{
				URL:        "try.gitea.io",
				Token:      "",
				Owner:      "olblak",
				Repository: "updatecli",
			},
			wantResult: "",
			wantErr:    true,
		},
		{
			name: "repository should exist with latest branch v3",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
			}{
				URL:        "try.gitea.io",
				Token:      "",
				Owner:      "olblak",
				Repository: "updatecli-test",
			},
			wantResult: "v3",
			wantErr:    false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			// Init gitea object
			g, gotErr := New(tt.manifest)
			require.NoError(t, gotErr)

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
