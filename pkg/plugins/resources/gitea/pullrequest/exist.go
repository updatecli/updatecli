package pullrequest

import (
	"context"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CheckActionExist verifies if an existing GitTea pullrequest is already opened.
func (g *Gitea) CheckActionExist(ctx context.Context, report *reports.Action) error {

	pullrequestTitle, pullrequestDescription, pullrequestLink, err := g.isPullRequestExist(ctx)
	if err != nil {
		return err
	}

	if pullrequestLink != "" {
		log.Debugf("GiTea pull request detected")

		report.Title = pullrequestTitle
		report.Link = pullrequestLink
		report.Description = pullrequestDescription
		return nil
	}

	return nil
}
