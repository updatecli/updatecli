package pullrequest

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreatePullRequest(t *testing.T) {

	// Disable source testing with running short tes
	if testing.Short() {
		return
	}

	tests := []struct {
		name     string
		manifest struct {
			URL          string
			Token        string
			Owner        string
			Repository   string
			SourceBranch string
			TargetBranch string
		}
		wantResult string
		wantErr    bool
	}{
		{
			name: "pullrequest shouldn't be created on olblak/updatecli should not exist",
			manifest: struct {
				URL          string
				Token        string
				Owner        string
				Repository   string
				SourceBranch string
				TargetBranch string
			}{
				URL:          "try.gitea.io",
				Token:        os.Getenv("UPDATECLI_GITEA_TOKEN"),
				Owner:        "olblak",
				Repository:   "updatecli-test",
				SourceBranch: "v1",
				TargetBranch: "main",
			},
			wantResult: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			g, gotErr := New(tt.manifest, nil)
			require.NoError(t, gotErr)

			gotErr = g.CreatePullRequest(
				"Bump version to x.y.z",
				"This is a changelog",
				"This is a report")

			if tt.wantErr {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}
		})

	}
}
