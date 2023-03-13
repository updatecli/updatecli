package mergerequest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gitlabclient "github.com/updatecli/updatecli/pkg/plugins/resources/gitlab/client"
	gitlabscm "github.com/updatecli/updatecli/pkg/plugins/scms/gitlab"
)

func TestNew(t *testing.T) {
	testData := []struct {
		name                       string
		spec                       Spec
		scm                        *gitlabscm.Gitlab
		expectedGitlabOwner        string
		expectedGitlabRepository   string
		expectedGitlabSourceBranch string
		expectedGitlabTargetBranch string
		wantErr                    bool
		wantErrMessage             string
	}{
		{
			name: "Test basic scenario",
			spec: Spec{
				Owner:        "updatecli",
				Repository:   "updatecli",
				SourceBranch: "workingBranch",
				TargetBranch: "main",
			},
			scm: &gitlabscm.Gitlab{
				Spec: gitlabscm.Spec{
					Spec: gitlabclient.Spec{
						URL:      "gitlab.com",
						Token:    "xxx",
						Username: "tes",
					},
				},
			},
			expectedGitlabOwner:        "updatecli",
			expectedGitlabRepository:   "updatecli",
			expectedGitlabSourceBranch: "workingBranch",
			expectedGitlabTargetBranch: "main",
		},
		{
			name: "Test parameter inheritance 1",
			spec: Spec{
				Owner:        "updatecli",
				Repository:   "updatecli",
				SourceBranch: "workingBranch",
				TargetBranch: "main",
			},
			scm: &gitlabscm.Gitlab{
				Spec: gitlabscm.Spec{
					Spec: gitlabclient.Spec{
						URL: "gitlab.com",
					},
					Repository: "updatecli-test",
					Branch:     "v2",
					Owner:      "tartempion",
				},
				HeadBranch: "workingBranch",
			},
			expectedGitlabOwner:        "updatecli",
			expectedGitlabRepository:   "updatecli",
			expectedGitlabSourceBranch: "workingBranch",
			expectedGitlabTargetBranch: "main",
		},
		{
			name: "Test parameter inheritance 2",
			spec: Spec{},
			scm: &gitlabscm.Gitlab{
				Spec: gitlabscm.Spec{
					Spec: gitlabclient.Spec{
						URL: "gitlab.com",
					},
					Repository: "updatecli-test",
					Branch:     "v2",
					Owner:      "tartempion",
				},
				HeadBranch: "workingBranchv2",
			},
			expectedGitlabOwner:        "tartempion",
			expectedGitlabRepository:   "updatecli-test",
			expectedGitlabSourceBranch: "workingBranchv2",
			expectedGitlabTargetBranch: "v2",
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

			assert.Equal(t, g.Owner, tt.expectedGitlabOwner)
			assert.Equal(t, g.Repository, tt.expectedGitlabRepository)
			assert.Equal(t, g.SourceBranch, tt.expectedGitlabSourceBranch)
			assert.Equal(t, g.TargetBranch, tt.expectedGitlabTargetBranch)
		})

	}

}
