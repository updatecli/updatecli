package pullrequest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	giteaclient "github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
	giteascm "github.com/updatecli/updatecli/pkg/plugins/scms/gitea"
)

func TestNew(t *testing.T) {
	testData := []struct {
		name                      string
		spec                      Spec
		scm                       *giteascm.Gitea
		expectedGiteaOwner        string
		expectedGiteaRepository   string
		expectedGiteaSourceBranch string
		expectedGiteaTargetBranch string
		wantErr                   bool
		wantErrMessage            string
	}{
		{
			name: "Test basic scenario",
			spec: Spec{
				Owner:        "updatecli",
				Repository:   "updatecli",
				SourceBranch: "workingBranch",
				TargetBranch: "main",
			},
			scm: &giteascm.Gitea{
				Spec: giteascm.Spec{
					Spec: giteaclient.Spec{
						URL:      "gitea.updatecli.io",
						Token:    "xxx",
						Username: "tes",
					},
				},
			},
			expectedGiteaOwner:        "updatecli",
			expectedGiteaRepository:   "updatecli",
			expectedGiteaSourceBranch: "workingBranch",
			expectedGiteaTargetBranch: "main",
		},
		{
			name: "Test basic scenario with assignees",
			spec: Spec{
				Owner:        "updatecli",
				Repository:   "updatecli",
				SourceBranch: "workingBranch",
				TargetBranch: "main",
				Assignees:    []string{"user1", "user2"},
			},
			scm: &giteascm.Gitea{
				Spec: giteascm.Spec{
					Spec: giteaclient.Spec{
						URL:      "gitea.updatecli.io",
						Token:    "xxx",
						Username: "tes",
					},
				},
			},
			expectedGiteaOwner:        "updatecli",
			expectedGiteaRepository:   "updatecli",
			expectedGiteaSourceBranch: "workingBranch",
			expectedGiteaTargetBranch: "main",
		},
		{
			name: "Test parameter inheritance 1",
			spec: Spec{
				Owner:        "updatecli",
				Repository:   "updatecli",
				SourceBranch: "workingBranch",
				TargetBranch: "main",
			},
			scm: &giteascm.Gitea{
				Spec: giteascm.Spec{
					Spec: giteaclient.Spec{
						URL: "gitea.updatecli.io",
					},
					Repository: "updatecli-test",
					Branch:     "v2",
					Owner:      "tartempion",
				},
			},
			expectedGiteaOwner:        "updatecli",
			expectedGiteaRepository:   "updatecli",
			expectedGiteaSourceBranch: "workingBranch",
			expectedGiteaTargetBranch: "main",
		},
		{
			name: "Test parameter inheritance 2",
			spec: Spec{},
			scm: &giteascm.Gitea{
				Spec: giteascm.Spec{
					Spec: giteaclient.Spec{
						URL: "gitea.updatecli.io",
					},
					Repository: "updatecli-test",
					Branch:     "v2",
					Owner:      "tartempion",
				},
			},
			expectedGiteaOwner:        "tartempion",
			expectedGiteaRepository:   "updatecli-test",
			expectedGiteaSourceBranch: "v2",
			expectedGiteaTargetBranch: "v2",
		},
		{
			name: "Test required parameter URL not specified",
			spec: Spec{
				Owner: "updatecli",
			},
			scm:     nil,
			wantErr: true,
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {

			g, gotErr := New(tt.spec, tt.scm)

			if tt.wantErr {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}

			assert.Equal(t, g.Owner, tt.expectedGiteaOwner)
			assert.Equal(t, g.Repository, tt.expectedGiteaRepository)
			assert.Equal(t, g.SourceBranch, tt.expectedGiteaSourceBranch)
			assert.Equal(t, g.TargetBranch, tt.expectedGiteaTargetBranch)
		})

	}

}
