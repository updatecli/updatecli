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
func (b *Bitbucket) CreateAction(report *reports.Action, resetDescription bool) error {
	title := report.Title
	if len(b.spec.Title) > 0 {
		title = b.spec.Title
	}

	// One Bitbucket pull request body can contain multiple action report
	// It would be better to refactor CreateAction to be able to reuse existing pull request description.
	// similar to what we did for github pull request.
	body, err := utils.GeneratePullRequestBodyMarkdown("", report.ToActionsMarkdownString())
	if err != nil {
		logrus.Warningf("something went wrong while generating Bitbucket Pull Request body: %s", err)
	}

	if len(b.spec.Body) > 0 {
		body = b.spec.Body
	}

	// Check if a pull-request is already opened then exit early if it does.
	pullrequestTitle, pullrequestDescription, pullrequestLink, err := b.isPullRequestExist()
	if err != nil {
		return err
	}

	if pullrequestLink != "" {
		logrus.Debugf("Bitbucket pull request already exist, nothing to do")

		report.Title = pullrequestTitle
		report.Link = pullrequestLink
		report.Description = pullrequestDescription
		return nil
	}

	// Test that both sourceBranch and targetBranch exists on remote before creating a new one
	ok, err := b.isRemoteBranchesExist()
	if err != nil {
		return err
	}

	/*
		Due to the following scenario, Updatecli always tries to open a Pull request
			* A pull request has been "manually" closed via UI
			* A previous Updatecli run failed during a Pull request creation for example due to network issues


		Therefore we always try to open a pull request, we don't consider being an error if all conditions are not met
		such as missing remote branches.
	*/

	if !ok {
		logrus.Debugln("skipping pull request creation")
		return nil
	}

	opts := scm.PullRequestInput{
		Title:  title,
		Body:   body,
		Source: b.SourceBranch,
		Target: b.TargetBranch,
	}

	logrus.Debugf("Title:\t%q\nBody:\t%q\nSource:\n%q\ntarget:\t%q\n",
		title,
		body,
		b.SourceBranch,
		b.TargetBranch)

	// Timeout api query after 30sec
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pr, resp, err := b.client.PullRequests.Create(
		ctx,
		strings.Join([]string{
			b.Owner,
			b.Repository,
		}, "/"),
		&opts,
	)

	if resp.Status > 400 {
		logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
	}

	if err != nil {
		if err.Error() == scm.ErrNotFound.Error() {
			logrus.Infof("Bitbucket Cloud pull request not created, skipping")
			return nil
		}
		return err
	}

	report.Link = pr.Link
	report.Description = pr.Body
	report.Title = pr.Title

	logrus.Infof("Bitbucket Cloud pull request successfully opened on %q", pr.Link)

	return nil
}
