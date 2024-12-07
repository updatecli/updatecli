package pullrequest

import (
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CheckActionExist verifies if an existing GitTea pullrequest is already opened.
func (g *Gitea) CheckActionExist(report *reports.Action) error {

	pullrequestTitle, pullrequestDescription, pullrequestLink, err := g.isPullRequestExist()
	if err != nil {
		return err
	}

	if pullrequestLink != "" {
		logrus.Debugf("GiTea pull request detected")

		report.Title = pullrequestTitle
		report.Link = pullrequestLink
		report.Description = pullrequestDescription
		return nil
	}

	return nil
}
