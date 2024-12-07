package pullrequest

import (
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CheckActionExist verifies if an existing BitBucket pullrequest is already opened
func (b *Bitbucket) CheckActionExist(report *reports.Action) error {

	pullrequestTitle, pullrequestDescription, pullrequestLink, err := b.isPullRequestExist()
	if err != nil {
		return err
	}

	if pullrequestLink != "" {
		logrus.Debugf("Bitbucket pull request detected")

		report.Title = pullrequestTitle
		report.Link = pullrequestLink
		report.Description = pullrequestDescription
		return nil
	}

	return nil
}
