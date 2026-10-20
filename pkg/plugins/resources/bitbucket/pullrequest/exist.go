package pullrequest

import (
	"context"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CheckActionExist verifies if an existing BitBucket pullrequest is already opened
func (b *Bitbucket) CheckActionExist(ctx context.Context, report *reports.Action) error {
	pullRequestExists, pullRequestDetails, err := b.isPullRequestExist()
	if err != nil {
		return err
	}

	if pullRequestExists {
		log.Debugf("Bitbucket pull request detected")

		report.Title = pullRequestDetails.Title
		report.Description = pullRequestDetails.Description
		report.Link = pullRequestDetails.Link
		return nil
	}

	return nil
}
