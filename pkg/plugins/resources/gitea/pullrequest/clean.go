package pullrequest

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CleanAction verifies if an existing action requires cleanup such as closing a pull request with no changes.
func (g *Gitea) CleanAction(ctx context.Context, report *reports.Action) error {
	existingPR, err := g.findExistingPullRequest(ctx)
	if err != nil {
		return fmt.Errorf("finding existing pull request: %w", err)
	}

	if existingPR == nil {
		log.Debugln("nothing to clean")
		return nil
	}

	isEmpty, err := g.isPullRequestEmpty(ctx, existingPR, cleanupMaxAttempts)
	if err != nil {
		return err
	}

	if !isEmpty {
		return nil
	}

	log.Debugf("No changed file detected at pull request:\n\t%s", existingPR.HTMLURL)

	if err := g.closePullRequest(ctx, existingPR.Index); err != nil {
		return fmt.Errorf("closing pull request: %w", err)
	}

	report.Link = ""
	report.Description = pullRequestClosedMessage

	return nil
}
