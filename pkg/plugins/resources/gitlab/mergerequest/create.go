package mergerequest

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/reports"

	utils "github.com/updatecli/updatecli/pkg/plugins/utils/action"
)

// CreateAction opens a Merge Request on the GitLab server
func (g *Gitlab) CreateAction(report *reports.Action, resetDescription bool) error {

	title := report.Title
	if len(g.spec.Title) > 0 {
		title = g.spec.Title
	}

	// One GitLab mergerequest body can contain multiple action report
	// It would be better to refactor CreateAction
	// to be able to reuse existing mergerequest description.
	// similar to what we did for github pullrequest.
	body, err := utils.GeneratePullRequestBody("", report.ToActionsString())
	if err != nil {
		logrus.Warningf("something went wrong while generating GitLab body: %s", err)
		return fmt.Errorf("generate GitLab body: %s", err.Error())
	}

	if len(g.spec.Body) > 0 {
		body = g.spec.Body
	}

	// Check if a merge-request is already opened then exit early if it does.
	mergeRequestTitle, mergeRequestDescription, mergeRequestLink, err := g.isMergeRequestExist()
	if err != nil {
		return fmt.Errorf("check if a mergerequest already exist: %s", err.Error())
	}

	// If a mergerequest already exist, we update the report with the existing mergerequest
	// At the moment Updatecli doesn't support updating an existing mergerequest
	if mergeRequestLink != "" {
		logrus.Debugln("GitLab mergerequest already exist, nothing to do")

		report.Title = mergeRequestTitle
		report.Link = mergeRequestLink
		report.Description = mergeRequestDescription
		return nil
	}

	// Test that both sourceBranch and targetBranch exists on remote before creating a new one
	ok, err := g.isRemoteBranchesExist()
	if err != nil {
		return fmt.Errorf("check if remote branches exist: %s", err.Error())
	}

	/*
				Due to the following scenario, Updatecli always tries to open a mergerequest
					* A mergerequest has been "manually" closed via UI
					* A previous Updatecli run failed during a mergerequest creation for example due to network issues

		isPullRequestExist
				Therefore we always try to open a mergerequest, we don't consider being an error if all conditions are not met
				such as missing remote branches.
	*/
	if !ok {
		return fmt.Errorf("remote branches %q and %q do not exist, we can't open a mergerequest", g.SourceBranch, g.TargetBranch)
	}

	opts := scm.PullRequestInput{
		Title:  title,
		Body:   body,
		Source: g.SourceBranch,
		Target: g.TargetBranch,
	}

	logrus.Debugf("Title:\t%q\nBody:\t%v\nSource branch:\t%q\nTarget branch:\t%q\n",
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

	if err != nil {
		if err.Error() == scm.ErrNotFound.Error() {
			logrus.Infof("GitLab pullrequest not created, skipping")
			return nil
		}
		logrus.Debugf("HTTP Response:\n\tReturn code: %d\n\n", resp.Status)

		return fmt.Errorf("create GitLab mergerequest: %v", err)
	}

	report.Link = pr.Link
	report.Title = pr.Title
	report.Description = pr.Body

	if resp.Status > 400 {
		logrus.Debugf("HTTP return code: %d\n\n", resp.Status)
	}

	logrus.Infof("GitLab mergerequest successfully opened on %q", pr.Link)

	return nil
}
