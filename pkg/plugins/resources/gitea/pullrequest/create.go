package pullrequest

import (
	giteasdk "code.gitea.io/sdk/gitea"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/reports"
	utils "github.com/updatecli/updatecli/pkg/plugins/utils/action"
)

// CreateAction opens a Pull Request on the Gitea server
func (g *Gitea) CreateAction(report *reports.Action, resetDescription bool) error {

	title := report.Title

	// One Gitea pullrequest body can contain multiple action report
	// It would be better to refactor CreateAction to be able to reuse existing pullrequest description.
	// similar to what we did for github pullrequest.
	body, err := utils.GeneratePullRequestBody("", report.ToActionsString())

	if err != nil {
		logrus.Warningf("something wrong happened while generating gitea pullrequest body: %s", err)
	}

	if len(g.spec.Body) > 0 {
		body = g.spec.Body
	}

	if len(g.spec.Title) > 0 {
		title = g.spec.Title
	}

	// Check if a pull-request is already opened then exit early if it does.
	pullrequestTitle, pullrequestDescription, pullrequestLink, err := g.isPullRequestExist()
	if err != nil {
		return err
	}

	// If a pullrequest already exist, we update the report with the existing pullrequest
	if pullrequestLink != "" {
		logrus.Debugf("Gitea pullrequest already exist, nothing to do")

		report.Title = pullrequestTitle
		report.Link = pullrequestLink
		report.Description = pullrequestDescription
		return nil
	}

	// Test that both sourceBranch and targetBranch exists on remote before creating a new one
	ok, err := g.isRemoteBranchesExist()

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
		logrus.Debugln("skipping pullrequest creation")
		return nil
	}

	logrus.Debugf("Title:\t%q\nBody:\t%q\nSource:\n%q\ntarget:\t%q\n",
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
		logrus.Debugf("Setting assignees for pull request: %v", g.spec.Assignees)
	}

	pr, resp, err := g.client.CreatePullRequest(g.Owner, g.Repository, sdkOpts)

	if resp != nil && resp.StatusCode > 400 {
		logrus.Debugf("RC: %s\nBody:\n%s", resp.Status, resp.Body)
	}

	if err != nil {
		logrus.Infof("Gitea pullrequest not created, skipping")
		return nil
	}

	report.Title = pr.Title
	report.Description = pr.Body
	report.Link = pr.URL

	logrus.Infof("Gitea pullrequest successfully opened on %q", pr.URL)

	return nil
}
