package pullrequest

import (
	"context"
	"fmt"

	giteasdk "gitea.dev/sdk"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/reports"
	utils "github.com/updatecli/updatecli/pkg/plugins/utils/action"
)

// CreateAction opens or updates a Pull Request on the Gitea server
func (g *Gitea) CreateAction(ctx context.Context, report *reports.Action, resetDescription bool) error {

	title := report.Title

	if len(g.spec.Title) > 0 {
		title = g.spec.Title
	}

	// Check if a pull-request is already opened then update it if it does.
	pullrequest, err := g.findExistingPullRequest(ctx)
	if err != nil {
		return err
	}

	// If a pullrequest already exist, we update it and the report with the merged description
	if pullrequest != nil {
		log.Debugf("Gitea pullrequest already exist, updating it")

		body, err := g.pullRequestBody(pullrequest.Body, report, resetDescription)
		if err != nil {
			return fmt.Errorf("generate Gitea pull request body: %w", err)
		}

		pr, _, err := g.client.PullRequests.EditPullRequest(
			ctx,
			g.Owner,
			g.Repository,
			pullrequest.Index,
			giteasdk.EditPullRequestOption{
				Title: title,
				Body:  &body,
			},
		)
		if err != nil {
			return fmt.Errorf("update Gitea pull request: %w", err)
		}

		report.Title = pr.Title
		report.Link = pr.HTMLURL
		report.Description = pr.Body
		return nil
	}

	body, err := g.pullRequestBody("", report, resetDescription)
	if err != nil {
		return fmt.Errorf("generate Gitea pull request body: %w", err)
	}

	// Test that both sourceBranch and targetBranch exists on remote before creating a new one
	ok, err := g.isRemoteBranchesExist(ctx)

	if err != nil {
		return err
	}

	/*
		Due to the following scenario, Updatecli always tries to open a Pullrequest
			* A pullrequest has been "manually" closed via UI
			* A previous Updatecli run failed during a Pullrequest creation for example due to network issues


		Therefore we always try to open a pull request, we don't consider being an error if all conditions are not met
		such as missing remote branches.
	*/
	if !ok {
		log.Debugln("skipping pullrequest creation")
		return nil
	}

	log.Debugf("Title:\t%q\nBody:\t%q\nSource:\n%q\ntarget:\t%q\n",
		title,
		body,
		g.SourceBranch,
		g.TargetBranch)

	sdkOpts := giteasdk.CreatePullRequestOption{
		Title:     title,
		Body:      body,
		Base:      g.TargetBranch,   // Base = Target branch
		Head:      g.SourceBranch,   // Head = Source branch
		Assignees: g.spec.Assignees, // Take list of assignees from spec
	}

	if len(g.spec.Assignees) > 0 {
		log.Debugf("Setting assignees for pull request: %v", g.spec.Assignees)
	}

	pr, resp, err := g.client.PullRequests.CreatePullRequest(ctx, g.Owner, g.Repository, sdkOpts)

	if resp != nil && resp.StatusCode > 400 {
		log.Debugf("RC: %s\nBody:\n%s", resp.Status, resp.Body)
	}

	if err != nil {
		log.Infof("Gitea pullrequest not created, skipping")
		return nil
	}

	report.Title = pr.Title
	report.Description = pr.Body
	report.Link = pr.HTMLURL

	log.Infof("Gitea pullrequest successfully opened on %q", pr.HTMLURL)

	return nil
}

// pullRequestBody generates the pull request description, merging the report
// with the existing description unless it must be reset.
func (g *Gitea) pullRequestBody(existingDescription string, report *reports.Action, resetBody bool) (string, error) {
	if len(g.spec.Body) > 0 {
		return g.spec.Body, nil
	}

	actions := report.ToActionsString()

	if resetBody {
		log.Debugf("Resetting pull request description with new report")
		return utils.GeneratePullRequestBody("", actions)
	}

	if existingDescription != "" {
		log.Debugf("Merging pull request description with new report")
		actions = reports.MergeFromString(existingDescription, actions)
	}

	return utils.GeneratePullRequestBody("", actions)
}
