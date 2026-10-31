package pullrequest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	giteasdk "gitea.dev/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

func TestCreateActionExistingPullRequestAutoMerge(t *testing.T) {
	fake := cleanServer{
		headRepoID: testRepoID,
		listedSha:  []string{"abc"},
		branchSha:  "abc",
	}

	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v1/version":
			_, _ = w.Write([]byte(`{"version":"1.24.0"}`))
		case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf("/api/v1/repos/updatecli/updatecli/pulls/%d/merge", testPRIndex):
			calls = append(calls, r.Method)
			w.WriteHeader(http.StatusOK)
		default:
			if r.Method == http.MethodPatch {
				calls = append(calls, r.Method)
			}
			fake.handler(w, r)
		}
	}))
	defer server.Close()

	client, err := giteasdk.NewClient(server.URL)
	require.NoError(t, err)

	g := Gitea{
		spec:         Spec{Merge: MergeSpec{Strategy: MergeStrategyAuto}, MergeMethod: "squash"},
		client:       client,
		Owner:        "updatecli",
		Repository:   "updatecli",
		SourceBranch: "updatecli",
		TargetBranch: "main",
	}

	report := reports.Action{Title: "title"}

	require.NoError(t, g.CreateAction(t.Context(), &report, false))

	// The auto-merge is requested once the pull request is up to date
	assert.Equal(t, []string{http.MethodPatch, http.MethodPost}, calls)
}

func TestCreateActionDraftPullRequest(t *testing.T) {
	testData := []struct {
		name          string
		remoteTitle   string
		draft         bool
		specTitle     string
		expectedTitle string
		expectedMerge bool
	}{
		{
			name:          "Pull request which isn't a draft follows the report",
			remoteTitle:   "bump app to 1.0.0",
			expectedTitle: "bump app to 2.0.0",
			expectedMerge: true,
		},
		{
			name:          "WIP prefix is kept",
			remoteTitle:   "WIP: bump app to 1.0.0",
			draft:         true,
			expectedTitle: "WIP: bump app to 2.0.0",
		},
		{
			name:          "Prefix is kept as it was written",
			remoteTitle:   "wip: bump app to 1.0.0",
			draft:         true,
			expectedTitle: "wip: bump app to 2.0.0",
		},
		{
			name:          "Bracket prefix is kept",
			remoteTitle:   "[WIP] bump app to 1.0.0",
			draft:         true,
			expectedTitle: "[WIP] bump app to 2.0.0",
		},
		{
			name:          "Prefix is detected without the draft field",
			remoteTitle:   "WIP: bump app to 1.0.0",
			expectedTitle: "WIP: bump app to 2.0.0",
		},
		{
			name:          "Unknown draft prefix keeps the title",
			remoteTitle:   "Draft: bump app to 1.0.0",
			draft:         true,
			expectedTitle: "Draft: bump app to 1.0.0",
		},
		{
			name:          "Manifest title with a prefix isn't prefixed twice",
			remoteTitle:   "WIP: bump app to 1.0.0",
			draft:         true,
			specTitle:     "WIP: custom title",
			expectedTitle: "WIP: custom title",
		},
		{
			name:          "Manifest title on a draft keeps the prefix",
			remoteTitle:   "WIP: bump app to 1.0.0",
			draft:         true,
			specTitle:     "custom title",
			expectedTitle: "WIP: custom title",
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			fake := cleanServer{
				headRepoID: testRepoID,
				listedSha:  []string{"abc"},
				branchSha:  "abc",
				title:      tt.remoteTitle,
				draft:      tt.draft,
			}

			merged := false
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case r.URL.Path == "/api/v1/version":
					_, _ = w.Write([]byte(`{"version":"1.24.0"}`))
				case r.Method == http.MethodPost && r.URL.Path == fmt.Sprintf("/api/v1/repos/updatecli/updatecli/pulls/%d/merge", testPRIndex):
					merged = true
					w.WriteHeader(http.StatusOK)
				default:
					fake.handler(w, r)
				}
			}))
			defer server.Close()

			client, err := giteasdk.NewClient(server.URL)
			require.NoError(t, err)

			g := Gitea{
				spec: Spec{
					Title:       tt.specTitle,
					Merge:       MergeSpec{Strategy: MergeStrategyAuto},
					MergeMethod: "squash",
				},
				client:       client,
				Owner:        "updatecli",
				Repository:   "updatecli",
				SourceBranch: "updatecli",
				TargetBranch: "main",
			}

			report := reports.Action{Title: "bump app to 2.0.0"}

			require.NoError(t, g.CreateAction(t.Context(), &report, false))

			assert.Equal(t, tt.expectedTitle, report.Title)
			assert.Equal(t, tt.expectedMerge, merged)
		})
	}
}
