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
func (s *Stash) CreateAction(report *reports.Action, resetDescription bool) error {

	title := report.Title
	if len(s.spec.Title) > 0 {
		title = s.spec.Title
	}

	// One Bitbucket pullrequest body can contain multiple action report
	// It would be better to refactor CreateAction to be able to reuse existing pullrequest description.
	// similar to what we did for github pullrequest.
	body, err := utils.GeneratePullRequestBodyMarkdown("", report.ToActionsMarkdownString())
	if err != nil {
		logrus.Warningf("something wrong happened while generating stash pullrequest body: %s", err)
	}

	if len(s.spec.Body) > 0 {
		body = s.spec.Body
	}

	// Check if a pull-request is already opened then exit early if it does.
	pullrequestTitle, pullrequestDescription, pullrequestLink, err := s.isPullRequestExist()
	if err != nil {
		return err
	}

	if pullrequestLink != "" {
		report.Title = pullrequestTitle
		report.Link = pullrequestLink
		report.Description = pullrequestDescription
		return nil
	}

	// Test that both sourceBranch and targetBranch exists on remote before creating a new one
	ok, err := s.isRemoteBranchesExist()

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
		Source: s.SourceBranch,
		Target: s.TargetBranch,
	}

	logrus.Debugf("Title:\t%q\nBody:\t%q\nSource:\n%q\ntarget:\t%q\n",
		title,
		body,
		s.SourceBranch,
		s.TargetBranch)

	// Timeout api query after 30sec
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pr, resp, err := s.client.PullRequests.Create(
		ctx,
		strings.Join([]string{
			s.Owner,
			s.Repository}, "/"),
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

	report.Link = pr.Link
	report.Title = pr.Title
	report.Description = pr.Body

	logrus.Infof("Bitbucket pullrequest successfully opened on %q", pr.Link)

	return nil
}
