package pullrequest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	giteasdk "gitea.dev/sdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/reports"
	utils "github.com/updatecli/updatecli/pkg/plugins/utils/action"
)

func testActionReport(targetID string) reports.Action {
	return reports.Action{
		ID:            "action",
		Title:         "report title",
		PipelineTitle: "pipeline",
		Targets: []reports.ActionTarget{
			{ID: targetID, Title: "title of " + targetID},
		},
	}
}

func TestCreateActionExistingPullRequest(t *testing.T) {
	previousReport := testActionReport("previous")
	previousBody, err := utils.GeneratePullRequestBody("", previousReport.ToActionsString())
	require.NoError(t, err)

	testData := []struct {
		name             string
		spec             Spec
		existingBody     string
		resetDescription bool
		patchStatus      int
		expectedError    bool
		expectedTitle    string
		expectedContains []string
		expectedMissing  []string
		expectedBody     string
	}{
		{
			name:             "Report is merged with the existing description",
			existingBody:     previousBody,
			expectedTitle:    "report title",
			expectedContains: []string{"title of previous", "title of current"},
		},
		{
			name:             "Report replaces the existing description on reset",
			existingBody:     previousBody,
			resetDescription: true,
			expectedTitle:    "report title",
			expectedContains: []string{"title of current"},
			expectedMissing:  []string{"title of previous"},
		},
		{
			name:          "Spec body is used unchanged",
			spec:          Spec{Body: "custom body"},
			existingBody:  previousBody,
			expectedTitle: "report title",
			expectedBody:  "custom body",
		},
		{
			name:             "Spec title is used",
			spec:             Spec{Title: "custom title"},
			existingBody:     previousBody,
			expectedTitle:    "custom title",
			expectedContains: []string{"title of previous", "title of current"},
		},
		{
			name:             "Description which isn't a report is replaced",
			existingBody:     "written by hand",
			expectedTitle:    "report title",
			expectedContains: []string{"title of current"},
			expectedMissing:  []string{"written by hand"},
		},
		{
			name:             "Empty description",
			expectedTitle:    "report title",
			expectedContains: []string{"title of current"},
		},
		{
			name:          "Failing to update is an error",
			existingBody:  previousBody,
			patchStatus:   http.StatusForbidden,
			expectedError: true,
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			fake := cleanServer{
				headRepoID:  testRepoID,
				listedSha:   []string{"abc"},
				branchSha:   "abc",
				title:       "remote title",
				body:        tt.existingBody,
				closeStatus: tt.patchStatus,
			}
			server := httptest.NewServer(http.HandlerFunc(fake.handler))
			defer server.Close()

			client, err := giteasdk.NewClient(server.URL)
			require.NoError(t, err)

			g := Gitea{
				spec:         tt.spec,
				client:       client,
				Owner:        "updatecli",
				Repository:   "updatecli",
				SourceBranch: "updatecli",
				TargetBranch: "main",
			}

			report := testActionReport("current")

			err = g.CreateAction(t.Context(), &report, tt.resetDescription)

			assert.Equal(t, 1, fake.closeCount)

			if tt.expectedError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			var sent giteasdk.EditPullRequestOption
			require.NoError(t, json.Unmarshal([]byte(fake.closeBody), &sent))
			require.NotNil(t, sent.Body)

			assert.Equal(t, tt.expectedTitle, sent.Title)
			assert.Nil(t, sent.State)
			assert.Nil(t, sent.Assignees)

			if tt.expectedBody != "" {
				assert.Equal(t, tt.expectedBody, *sent.Body)
			}
			for _, expected := range tt.expectedContains {
				assert.Contains(t, *sent.Body, expected)
			}
			for _, missing := range tt.expectedMissing {
				assert.NotContains(t, *sent.Body, missing)
			}

			// The report reflects the pull request returned by the API
			assert.Equal(t, tt.expectedTitle, report.Title)
			assert.Equal(t, *sent.Body, report.Description)
			assert.Equal(t, "https://gitea.example.com/updatecli/updatecli/pulls/42", report.Link)
		})
	}
}

func TestCreateActionForkPullRequest(t *testing.T) {
	const openedIndex = testPRIndex + 1
	openedLink := "https://gitea.example.com/updatecli/updatecli/pulls/43"
	pullsPath := "/api/v1/repos/updatecli/updatecli/pulls"
	openedPath := fmt.Sprintf("%s/%d", pullsPath, openedIndex)

	testData := []struct {
		name          string
		requestStatus int
	}{
		{
			name:          "Reviewer request succeeds",
			requestStatus: http.StatusCreated,
		},
		{
			name:          "Reviewer request fails",
			requestStatus: http.StatusForbidden,
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			fake := cleanServer{
				headRepoID: testForkID,
				listedSha:  []string{"abc"},
				branchSha:  "abc",
				title:      "remote title",
				body:       "remote body",
			}

			var calls []string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				switch {
				case r.URL.Path == "/api/v1/version":
					_, _ = w.Write([]byte(`{"version":"1.24.0"}`))
				case r.Method == http.MethodGet && r.URL.Path == "/api/v1/settings/api":
					_, _ = w.Write([]byte(`{"max_response_items":50}`))
				case r.Method == http.MethodGet && r.URL.Path == "/api/v1/repos/updatecli/updatecli/branches":
					_ = json.NewEncoder(w).Encode([]giteasdk.Branch{{Name: "updatecli"}, {Name: "main"}})
				case r.Method == http.MethodPost && r.URL.Path == pullsPath:
					calls = append(calls, r.URL.Path)
					w.WriteHeader(http.StatusCreated)
					_ = json.NewEncoder(w).Encode(giteasdk.PullRequest{
						Index:   openedIndex,
						HTMLURL: openedLink,
						Title:   "report title",
						Head:    &giteasdk.PRBranchInfo{Name: "updatecli", Sha: "def", RepoID: testRepoID},
						Base:    &giteasdk.PRBranchInfo{Name: "main", RepoID: testRepoID},
					})
				case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/merge"):
					calls = append(calls, r.URL.Path)
				case r.Method == http.MethodGet && r.URL.Path == openedPath+"/reviews":
					w.Header().Set("X-Total-Count", "0")
					_, _ = w.Write([]byte(`[]`))
				case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/requested_reviewers"):
					calls = append(calls, r.URL.Path)
					w.WriteHeader(tt.requestStatus)
				default:
					fake.handler(w, r)
				}
			}))
			defer server.Close()

			client, err := giteasdk.NewClient(server.URL)
			require.NoError(t, err)

			g := Gitea{
				spec: Spec{
					Owner:        "updatecli",
					Repository:   "updatecli",
					SourceBranch: "updatecli",
					TargetBranch: "main",
					Merge:        MergeSpec{Strategy: MergeStrategyAuto},
					MergeMethod:  "squash",
					Reviewers:    []string{"alice"},
				},
				client:       client,
				Owner:        "updatecli",
				Repository:   "updatecli",
				SourceBranch: "updatecli",
				TargetBranch: "main",
			}

			report := testActionReport("current")

			err = g.CreateAction(t.Context(), &report, false)
			if tt.requestStatus == http.StatusCreated {
				require.NoError(t, err)
				assert.Equal(t, []string{pullsPath, openedPath + "/requested_reviewers", openedPath + "/merge"}, calls)
			} else {
				require.ErrorContains(t, err, "requesting reviewers")
				assert.Equal(t, []string{pullsPath, openedPath + "/requested_reviewers"}, calls)
			}

			// The fork's pull request is left untouched and Updatecli opens its own
			assert.Zero(t, fake.closeCount)
			assert.Equal(t, openedLink, report.Link)
		})
	}
}

func TestCheckActionExistForkPullRequest(t *testing.T) {
	fake := cleanServer{
		headRepoID: testForkID,
		listedSha:  []string{"abc"},
		title:      "remote title",
		body:       "remote body",
	}
	server := httptest.NewServer(http.HandlerFunc(fake.handler))
	defer server.Close()

	client, err := giteasdk.NewClient(server.URL)
	require.NoError(t, err)

	g := Gitea{
		client:       client,
		Owner:        "updatecli",
		Repository:   "updatecli",
		SourceBranch: "updatecli",
		TargetBranch: "main",
	}

	report := reports.Action{}

	require.NoError(t, g.CheckActionExist(t.Context(), &report))

	assert.Empty(t, report.Link)
	assert.Empty(t, report.Title)
}

func TestCreateActionFailure(t *testing.T) {
	testData := []struct {
		name string
		spec Spec
	}{
		{
			name: "Without auto-merge",
		},
		{
			name: "With auto-merge",
			spec: Spec{Merge: MergeSpec{Strategy: MergeStrategyAuto}},
		},
	}

	for _, tt := range testData {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				switch {
				case r.Method == http.MethodGet && r.URL.Path == "/api/v1/repos/updatecli/updatecli/pulls":
					_ = json.NewEncoder(w).Encode([]giteasdk.PullRequest{})
				case r.Method == http.MethodGet && r.URL.Path == "/api/v1/repos/updatecli/updatecli/branches":
					_ = json.NewEncoder(w).Encode([]giteasdk.Branch{{Name: "updatecli"}, {Name: "main"}})
				case r.Method == http.MethodPost && r.URL.Path == "/api/v1/repos/updatecli/updatecli/pulls":
					http.Error(w, `{"message":"forbidden"}`, http.StatusForbidden)
				default:
					http.Error(w, "unexpected request "+r.Method+" "+r.URL.Path, http.StatusNotFound)
				}
			}))
			defer server.Close()

			client, err := giteasdk.NewClient(server.URL)
			require.NoError(t, err)

			tt.spec.Owner = "updatecli"
			tt.spec.Repository = "updatecli"
			tt.spec.SourceBranch = "updatecli"
			tt.spec.TargetBranch = "main"

			g := Gitea{
				spec:         tt.spec,
				client:       client,
				Owner:        "updatecli",
				Repository:   "updatecli",
				SourceBranch: "updatecli",
				TargetBranch: "main",
			}

			report := testActionReport("current")

			err = g.CreateAction(t.Context(), &report, false)

			require.ErrorContains(t, err, "create Gitea pull request")
			assert.Empty(t, report.Link)
		})
	}
}
