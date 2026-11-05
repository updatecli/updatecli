package pullrequest

import (
	"context"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CheckActionExist verifies if an existing GitTea pullrequest is already opened.
func (g *Gitea) CheckActionExist(ctx context.Context, report *reports.Action) error {

	pullrequest, err := g.findExistingPullRequest(ctx)
	if err != nil {
		return err
	}

	if pullrequest != nil {
		log.Debugf("GiTea pull request detected")

		report.Title = pullrequest.Title
		report.Link = pullrequest.HTMLURL
		report.Description = pullrequest.Body
		return nil
	}

	return nil
}
