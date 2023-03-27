package pullrequest

import (
	"context"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/reports"
	utils "github.com/updatecli/updatecli/pkg/plugins/utils/action"
)

// CreateAction opens a Pull Request on the Bitbucket server
func (g *Stash) CreateAction(report reports.Action) error {

	title := report.Title
	if len(g.spec.Title) > 0 {
		title = g.spec.Title
	}

	body, err := utils.GeneratePullRequestBody("", report.ToActionsString())
	if err != nil {
		logrus.Warningf("something wrong happened while generating stash pullrequest body: %s", err)
	}

	if len(g.spec.Body) > 0 {
		body = g.spec.Body
	}

	// Check if a pull-request is already opened then exit early if it does.
	exist, err := g.isPullRequestExist()
	if err != nil {
		return err
	}

	if exist {
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

	opts := scm.PullRequestInput{
		Title:  title,
		Body:   body,
		Source: g.SourceBranch,
		Target: g.TargetBranch,
	}

	logrus.Debugf("Title:\t%q\nBody:\t%q\nSource:\n%q\ntarget:\t%q\n",
		title,
		body,
		g.SourceBranch,
		g.TargetBranch)

	// Timeout api query after 30sec
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pr, resp, err := g.client.PullRequests.Create(
		ctx,
		strings.Join([]string{
			g.Owner,
			g.Repository}, "/"),
		&opts,
	)

	if resp.Status > 400 {
		logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
	}

	if err != nil {
		if err.Error() == scm.ErrNotFound.Error() {
			logrus.Infof("Bitbucket pullrequest not created, skipping")
			return nil
		}
		return err
	}

	logrus.Infof("Bitbucket pullrequest successfully opened on %q", pr.Link)

	return nil
}
