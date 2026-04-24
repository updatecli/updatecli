package pullrequest

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CheckActionExist verifies if an Azure DevOps pull request is already opened.
func (a *AzureDevOps) CheckActionExist(ctx context.Context, report *reports.Action) error {
	pr, err := a.findExistingPullRequest(ctx)
	if err != nil {
		return err
	}

	if pr != nil {
		logrus.Debugf("Azure DevOps pull request detected")

		report.Title = stringValue(pr.Title)
		report.Link = a.pullRequestLink(pr)
		report.Description = stringValue(pr.Description)
	}

	return nil
}
