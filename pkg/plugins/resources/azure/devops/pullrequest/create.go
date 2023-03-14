package pullrequest

import (
	"context"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/sirupsen/logrus"
)

// CreateAction opens a Pull Request on the Azure DevOps server
func (a *Azure) CreateAction(title, changelog, pipelineReport string) error {

	body := changelog + "\n" + pipelineReport

	if len(a.spec.Body) > 0 {
		body = a.spec.Body
	}

	if len(a.spec.Title) > 0 {
		title = a.spec.Title
	}

	// Check if a pull-request is already opened then exit early if it does.
	exist, err := a.isPullRequestExist()
	if err != nil {
		return err
	}

	if exist {
		return nil
	}

	// Test that both sourceBranch and targetBranch exists on remote before creating a new one
	ok, err := a.isRemoteBranchesExist()

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
		Source: a.SourceBranch,
		Target: a.TargetBranch,
	}

	logrus.Debugf("Title:\t%q\nBody:\t%q\nSource:\n%q\ntarget:\t%q\n",
		title,
		body,
		a.SourceBranch,
		a.TargetBranch)

	// Timeout api query after 30sec
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	pr, resp, err := a.client.PullRequests.Create(
		ctx,
		a.spec.RepoID,
		&opts,
	)

	if resp.Status > 400 {
		logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
	}

	if err != nil {
		if err.Error() == scm.ErrNotFound.Error() {
			logrus.Infof("Azure DevOps pullrequest not created, skipping")
			return nil
		}
		return err
	}

	logrus.Infof("Azure DevOps pullrequest successfully opened on %q", pr.Link)

	return nil
}
