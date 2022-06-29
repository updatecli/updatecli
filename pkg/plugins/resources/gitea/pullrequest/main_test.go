package pullrequest

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
)

func TestSource(t *testing.T) {

	tests := []struct {
		name       string
		spec       Spec
		wantResult string
		wantErr    bool
	}{
		{
			name: "pullrequest shouldn't be created on olblak/updatecli should not exist",
			spec: Spec{
				Spec: client.Spec{
					URL:   "try.gitea.io",
					Token: "",
				},
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

			g, _ := New(tt.spec)
			gotErr := g.CreatePullRequest(
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
