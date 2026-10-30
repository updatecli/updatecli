package pullrequest

import (
	"context"
	"fmt"
	"strings"

	giteasdk "gitea.dev/sdk"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/reports"
	utils "github.com/updatecli/updatecli/pkg/plugins/utils/action"
)

// CreateAction opens a Pull Request on the Gitea server
func (g *Gitea) CreateAction(ctx context.Context, report *reports.Action, resetDescription bool) error {

	title := report.Title

	// One Gitea pullrequest body can contain multiple action report
	// It would be better to refactor CreateAction to be able to reuse existing pullrequest description.
	// similar to what we did for github pullrequest.
	body, err := utils.GeneratePullRequestBody("", report.ToActionsString())

	if err != nil {
		log.Warningf("something wrong happened while generating gitea pullrequest body: %s", err)
	}

	if len(g.spec.Body) > 0 {
		body = g.spec.Body
	}

	if len(g.spec.Title) > 0 {
		title = g.spec.Title
	}

	// Check if a pull-request is already opened then exit early if it does.
	pullrequest, err := g.findExistingPullRequest(ctx)
	if err != nil {
		return err
	}

	// If a pullrequest already exist, we update the report with the existing pullrequest
	if pullrequest != nil {
		// A pull request opened from a fork isn't managed by Updatecli, even if it uses the same branch name
		if pullrequest.Base == nil || pullrequest.Head.RepoID != pullrequest.Base.RepoID {
			log.Debugf("Skipping Gitea pull request auto-merge as its head is on another repository:\n\t%s", pullrequest.URL)

			report.Title = pullrequest.Title
			report.Link = pullrequest.URL
			report.Description = pullrequest.Body
			return nil
		}

		log.Debugf("Gitea pullrequest already exists")

		report.Title = pullrequest.Title
		report.Link = pullrequest.URL
		report.Description = pullrequest.Body
		return g.enableAutoMerge(ctx, pullrequest)
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
		if strings.EqualFold(g.spec.Merge.Strategy, MergeStrategyAuto) {
			return fmt.Errorf("creating Gitea pull request for auto-merge: %w", err)
		}
		log.Infof("Gitea pullrequest not created, skipping")
		return nil
	}

	report.Title = pr.Title
	report.Description = pr.Body
	report.Link = pr.URL

	log.Infof("Gitea pullrequest successfully opened on %q", pr.URL)

	return g.enableAutoMerge(ctx, pr)
}
