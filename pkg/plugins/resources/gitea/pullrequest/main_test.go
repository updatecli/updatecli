package pullrequest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	giteaclient "github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/credential"
	giteascm "github.com/updatecli/updatecli/pkg/plugins/scms/gitea"
)

func TestNewTokenPrecedence(t *testing.T) {
	testData := []struct {
		name     string
		scmToken string
		expected string
	}{
		{
			name:     "Scm token wins over the environment",
			scmToken: "scm-token",
			expected: "token scm-token",
		},
		{
			name:     "Environment token is used without a scm token",
			expected: "token env-token",
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(credential.ENVIRONMENT_GITEA_TOKEN, "env-token")

			var authorization string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				authorization = r.Header.Get("Authorization")
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte("[]"))
			}))
			defer server.Close()

			g, err := New(Spec{}, &giteascm.Gitea{
				Spec: giteascm.Spec{
					Spec:       giteaclient.Spec{URL: server.URL, Token: tt.scmToken},
					Owner:      "updatecli",
					Repository: "updatecli",
					Branch:     "main",
				},
			})
			require.NoError(t, err)

			_, err = g.findExistingPullRequest(t.Context())
			require.NoError(t, err)

			assert.Equal(t, tt.expected, authorization)
		})
	}
}

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
		{
			name:           "Unsupported merge strategy",
			spec:           Spec{Merge: MergeSpec{Strategy: "client"}},
			wantErr:        true,
			wantErrMessage: `invalid merge strategy "client"`,
		},
		{
			name:           "Unsupported merge method",
			spec:           Spec{MergeMethod: "fast-forward"},
			wantErr:        true,
			wantErrMessage: `invalid merge method "fast-forward"`,
		},
	}

	for _, tt := range testData {

		t.Run(tt.name, func(t *testing.T) {

			g, gotErr := New(tt.spec, tt.scm)

			if tt.wantErr {
				require.Error(t, gotErr)
				if tt.wantErrMessage != "" {
					assert.ErrorContains(t, gotErr, tt.wantErrMessage)
				}
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
