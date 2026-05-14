package pullrequest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	azdoclient "github.com/updatecli/updatecli/pkg/plugins/resources/azuredevops/client"
	azdoscm "github.com/updatecli/updatecli/pkg/plugins/scms/azuredevops"
)

func TestNew(t *testing.T) {
	testData := []struct {
		name                 string
		spec                 Spec
		scm                  *azdoscm.AzureDevOps
		expectedProject      string
		expectedRepository   string
		expectedSourceBranch string
		expectedTargetBranch string
		wantErr              bool
	}{
		{
			name: "Test basic scenario",
			spec: Spec{
				SourceBranch: "workingBranch",
				TargetBranch: "main",
				Spec: azdoclient.Spec{
					URL:          "https://dev.azure.com",
					Organization: "updatecli",
					Project:      "updatecli",
					Repository:   "updatecli",
				},
			},
			expectedProject:      "updatecli",
			expectedRepository:   "updatecli",
			expectedSourceBranch: "workingBranch",
			expectedTargetBranch: "main",
		},
		{
			name: "Test parameter inheritance",
			spec: Spec{},
			scm: &azdoscm.AzureDevOps{
				Spec: azdoscm.Spec{
					Branch: "main",
					Spec: azdoclient.Spec{
						URL:          "https://dev.azure.com",
						Organization: "updatecli",
						Project:      "updatecli-project",
						Repository:   "updatecli-repository",
					},
				},
			},
			expectedProject:      "updatecli-project",
			expectedRepository:   "updatecli-repository",
			expectedSourceBranch: "main",
			expectedTargetBranch: "main",
		},
		{
			name: "Test default URL when not specified",
			spec: Spec{
				Spec: azdoclient.Spec{
					Organization: "updatecli",
					Project:      "updatecli",
					Repository:   "updatecli",
				},
			},
			expectedProject:    "updatecli",
			expectedRepository: "updatecli",
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			g, gotErr := New(tt.spec, tt.scm)

			if tt.wantErr {
				require.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, g.Project, tt.expectedProject)
			assert.Equal(t, g.Repository, tt.expectedRepository)
			assert.Equal(t, g.SourceBranch, tt.expectedSourceBranch)
			assert.Equal(t, g.TargetBranch, tt.expectedTargetBranch)
		})
	}
}

func TestRefName(t *testing.T) {
	testData := []struct {
		name     string
		branch   string
		expected string
	}{
		{
			name:     "Plain branch",
			branch:   "main",
			expected: "refs/heads/main",
		},
		{
			name:     "Already full ref",
			branch:   "refs/heads/main",
			expected: "refs/heads/main",
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, refName(tt.branch))
		})
	}
}
