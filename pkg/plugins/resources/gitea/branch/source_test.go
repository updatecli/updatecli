package branch

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			},
			wantResult: "updatecli_012c5c9cec4df4969325e5d428775a5b7d9e6726cb6a9ef0402f61ea102cd2b3",
			wantErr:    false,
		},
		{
			name: "repository should exist with latest branch v3",
			spec: Spec{
				Spec: client.Spec{
					URL:   "try.gitea.io",
					Token: "",
				},
				Owner:      "olblak",
				Repository: "updatecli-test",
			},
			wantResult: "v3",
			wantErr:    false,
		},
		//{
		//	name: "repository should exist with no branch v1",
		//	spec: Spec{
		//		Spec: client.Spec{
		//			URL:   "try.gitea.io",
		//			Token: "",
		//		},
		//		Owner:      "olblak",
		//		Repository: "updatecli-test",
		//		VersionFilter: version.Filter{
		//			Kind:    "regex",
		//			Pattern: "v1",
		//		},
		//	},
		//	wantResult: "v1",
		//	wantErr:    false,
		//},
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
