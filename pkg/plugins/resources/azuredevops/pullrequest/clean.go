package pullrequest

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/reports"
)

// CleanAction verifies if existing action requires cleanup such as closing a pull request with no changes.
func (a *AzureDevOps) CleanAction(ctx context.Context, report *reports.Action) error {
	existingPR, err := a.findExistingPullRequest(ctx)
	if err != nil {
		return fmt.Errorf("finding existing pull request: %w", err)
	}

	if existingPR == nil || existingPR.PullRequestId == nil {
		log.Debugln("nothing to clean")
		return nil
	}

	repository, err := a.client.GetRepository(ctx, a.Project, a.Repository)
	if err != nil {
		return fmt.Errorf("get Azure DevOps repository: %w", err)
	}

	repositoryID, err := repositoryID(repository)
	if err != nil {
		return err
	}

	gitClient, err := a.client.NewGitClient(ctx)
	if err != nil {
		return fmt.Errorf("create Azure DevOps git client: %w", err)
	}

	headMatches, err := a.retryUntilPullRequestHeadMatchesRemoteBranchHead(ctx, gitClient, repositoryID, *existingPR.PullRequestId, 3)
	if err != nil {
		return err
	}

	if !headMatches {
		log.Debugf("Skipping Azure DevOps pull request cleanup because PR head does not match remote branch head:\n\t%s", a.pullRequestLink(existingPR))
		return nil
	}

	isEmpty, err := a.isPullRequestEmpty(ctx, gitClient, repositoryID, *existingPR.PullRequestId)
	if err != nil {
		return err
	}

	if !isEmpty {
		return nil
	}

	log.Debugf("No changed file detected at pull request:\n\t%s", a.pullRequestLink(existingPR))

	if err := a.closePullRequest(ctx, gitClient, repositoryID, *existingPR.PullRequestId); err != nil {
		return fmt.Errorf("closing pull request: %w", err)
	}

	report.Link = ""
	report.Description = "pull request closed as no changed file detected"

	return nil
}
