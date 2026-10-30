package pullrequest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
