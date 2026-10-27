package pullrequest

import (
	"context"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CheckActionExist verifies if an existing GitTea pullrequest is already opened.
func (g *Gitea) CheckActionExist(ctx context.Context, report *reports.Action) error {

	pr, err := g.isPullRequestExist(ctx)
	if err != nil {
		return err
	}

	if pr != nil {
		log.Debugf("GiTea pull request detected")

		report.Title = pr.Title
		report.Link = pr.URL
		report.Description = pr.Body
		return nil
	}

	return nil
}
