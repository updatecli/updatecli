package pullrequest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	azureclient "github.com/updatecli/updatecli/pkg/plugins/resources/azure/devops/client"
	azurescm "github.com/updatecli/updatecli/pkg/plugins/scms/azure"
)

func TestNew(t *testing.T) {
	testData := []struct {
		name                 string
		spec                 Spec
		scm                  *azurescm.Azure
		expectedOwner        string
		expectedProject      string
		expectedRepoID       string
		expectedSourceBranch string
		expectedTargetBranch string
		wantErr              bool
		wantErrMessage       string
	}{
		{
			name: "Test basic scenario",
			spec: Spec{
				Spec: azureclient.Spec{
					Owner:   "updatecli",
					Project: "updatecli-test",
					RepoID:  "updatecli-test",
				},
				SourceBranch: "workingBranch",
				TargetBranch: "main",
			},
			scm: &azurescm.Azure{
				Spec: azurescm.Spec{
					Spec: azureclient.Spec{
						URL:     "dev.azure.com",
						Owner:   "updatecli",
						Project: "updatecli-test",
						RepoID:  "updatecli-test",
					},
					Branch: "v2",
				},
				HeadBranch: "workingBranch",
			},
			expectedOwner:        "updatecli",
			expectedProject:      "updatecli-test",
			expectedRepoID:       "updatecli-test",
			expectedSourceBranch: "workingBranch",
			expectedTargetBranch: "main",
		},
		{
			name: "Test parameter inheritance 1",
			spec: Spec{
				Spec: azureclient.Spec{
					Owner:   "updatecli",
					Project: "updatecli-test",
					RepoID:  "updatecli-test",
				},
				SourceBranch: "workingBranch",
				TargetBranch: "main",
			},
			scm: &azurescm.Azure{
				Spec: azurescm.Spec{
					Spec: azureclient.Spec{
						URL:     "dev.azure.com",
						Owner:   "updatecli",
						Project: "updatecli",
						RepoID:  "updatecli",
					},
					Branch: "v2",
				},
				HeadBranch: "workingBranch",
			},
			expectedOwner:        "updatecli",
			expectedProject:      "updatecli",
			expectedRepoID:       "updatecli",
			expectedSourceBranch: "workingBranch",
			expectedTargetBranch: "main",
		},
		{
			name: "Test parameter inheritance 2",
			spec: Spec{},
			scm: &azurescm.Azure{
				Spec: azurescm.Spec{
					Spec: azureclient.Spec{
						Owner:   "updatecli",
						Project: "updatecli-test",
						RepoID:  "updatecli-test",
					},
					Branch: "v2",
				},
				HeadBranch: "workingBranchv2",
			},
			expectedOwner:        "updatecli",
			expectedProject:      "updatecli-test",
			expectedRepoID:       "updatecli-test",
			expectedSourceBranch: "workingBranchv2",
			expectedTargetBranch: "v2",
		},
		{
			name: "Test required parameter URL not specified",
			spec: Spec{
				Spec: azureclient.Spec{
					Owner: "updatecli",
				},
				SourceBranch: "workingBranch",
				TargetBranch: "main",
			},
			scm:     nil,
			wantErr: true,
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {

			a, gotErr := New(tt.spec, tt.scm)

			if tt.wantErr {
				require.Error(t, gotErr)
			} else {
				require.NoError(t, gotErr)
			}

			assert.Equal(t, a.spec.Owner, tt.expectedOwner)
			assert.Equal(t, a.spec.Project, tt.expectedProject)
			assert.Equal(t, a.spec.RepoID, tt.expectedRepoID)
			assert.Equal(t, a.SourceBranch, tt.expectedSourceBranch)
			assert.Equal(t, a.TargetBranch, tt.expectedTargetBranch)
		})
	}
}
