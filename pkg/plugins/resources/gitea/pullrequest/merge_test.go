package pullrequest

import (
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
