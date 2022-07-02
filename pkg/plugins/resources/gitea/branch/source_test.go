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
			name: "repository olblak/updatecli-mirror should exist but no release",
			manifest: struct {
				URL        string
				Token      string
				Owner      string
				Repository string
			}{
				URL:        "try.gitea.io",
				Token:      "",
				Owner:      "olblak",
				Repository: "updatecli-mirror",
			},
			wantResult: "updatecli_012c5c9cec4df4969325e5d428775a5b7d9e6726cb6a9ef0402f61ea102cd2b3",
			wantErr:    false,
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
			g, _ := New(tt.manifest)

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
