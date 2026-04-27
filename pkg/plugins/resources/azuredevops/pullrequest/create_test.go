package pullrequest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/reports"
	utils "github.com/updatecli/updatecli/pkg/plugins/utils/action"
)

func TestPullRequestBody(t *testing.T) {
	t.Run("uses custom body when specified", func(t *testing.T) {
		pr := AzureDevOps{
			spec: Spec{
				Body: "custom body",
			},
		}

		body, err := pr.pullRequestBody("existing body", &reports.Action{}, false)

		require.NoError(t, err)
		assert.Equal(t, "custom body", body)
	})

	t.Run("generates markdown body for new pull requests", func(t *testing.T) {
		report := &reports.Action{
			Title:       "test",
			Description: "update description",
		}
		pr := AzureDevOps{}

		got, err := pr.pullRequestBody("", report, false)
		require.NoError(t, err)

		expected, err := utils.GeneratePullRequestBodyMarkdown("", report.ToActionsMarkdownString())
		require.NoError(t, err)

		assert.Equal(t, expected, got)
	})

	t.Run("merges existing markdown body for pull request updates", func(t *testing.T) {
		report := &reports.Action{
			Title:       "test",
			Description: "new update description",
		}
		pr := AzureDevOps{}

		got, err := pr.pullRequestBody("existing body", report, true)
		require.NoError(t, err)

		mergedDescription, err := reports.MergeFromMarkdown("existing body", report.ToActionsMarkdownString())
		require.NoError(t, err)

		expected, err := utils.GeneratePullRequestBodyMarkdown("", mergedDescription)
		require.NoError(t, err)

		assert.Equal(t, expected, got)
	})
}
