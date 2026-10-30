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

func TestCreateActionForkPullRequestSkipsAutoMerge(t *testing.T) {
	// The fake server rejects merge requests so requesting an auto-merge returns an error
	fake := cleanServer{
		headRepoID: testForkID,
		listedSha:  []string{"abc"},
		branchSha:  "abc",
	}
	server := httptest.NewServer(http.HandlerFunc(fake.handler))
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

	report := reports.Action{}

	require.NoError(t, g.CreateAction(t.Context(), &report, false))

	assert.Equal(t, 1, fake.listCount)
	assert.Zero(t, fake.closeCount)
	assert.Equal(t, "https://gitea.example.com/updatecli/updatecli/pulls/42", report.Link)
}

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
