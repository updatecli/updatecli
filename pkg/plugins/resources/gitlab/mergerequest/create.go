package mergerequest

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/reports"

	utils "github.com/updatecli/updatecli/pkg/plugins/utils/action"
	gitlabapi "gitlab.com/gitlab-org/api/client-go"
)

const gitlabRequestTimeout = 30 * time.Second

// CreateAction opens a Merge Request on the GitLab server
func (g *Gitlab) CreateAction(report *reports.Action, resetDescription bool) error {

	ctx, cancel := context.WithTimeout(context.Background(), gitlabRequestTimeout)
	defer cancel()

	labelOptions := gitlabapi.LabelOptions(g.spec.Labels)

	acceptMergeRequestOptions := gitlabapi.AcceptMergeRequestOptions{
		MergeWhenPipelineSucceeds: g.spec.MergeWhenPipelineSucceeds,
		MergeCommitMessage:        g.spec.MergeCommitMessage,
		SquashCommitMessage:       g.spec.SquashCommitMessage,
		Squash:                    g.spec.Squash,
		ShouldRemoveSourceBranch:  g.spec.RemoveSourceBranch,
	}

	var body string
	title := report.Title

	if len(g.spec.Title) > 0 {
		title = g.spec.Title
	}

	// Check if a merge-request is already opened then exit early if it does.
	existingMR, err := g.findExistingMR()
	if err != nil {
		return fmt.Errorf("check if a mergerequest already exist: %s", err.Error())
	}

	// If a mergerequest already exist, we update the report with the existing mergerequest
	if existingMR != nil {
		logrus.Debugln("GitLab mergerequest already exist, updating it")

		mergedDescription := reports.MergeFromString(existingMR.Description, report.ToActionsString())
		body, err = utils.GeneratePullRequestBody("", mergedDescription)
		if err != nil {
			logrus.Warningf("something went wrong while generating GitLab body: %s", err)
			return fmt.Errorf("generate GitLab body: %s", err.Error())
		}

		report.Title = existingMR.Title
		report.Link = existingMR.WebURL
		report.Description = body

		opts := &gitlabapi.UpdateMergeRequestOptions{
			Description:        &body,
			Title:              &title,
			AssigneeIDs:        &g.spec.Assignees,
			ReviewerIDs:        &g.spec.Reviewers,
			Squash:             g.spec.Squash,
			RemoveSourceBranch: g.spec.RemoveSourceBranch,
			Labels:             &labelOptions,
			AllowCollaboration: g.spec.AllowCollaboration,
		}

		_, _, err := g.api.MergeRequests.UpdateMergeRequest(
			g.getPID(),
			existingMR.IID,
			opts,
			gitlabapi.WithContext(ctx),
		)

		if err != nil {
			logrus.Warningf("something went wrong updating gitlab merge request %s/%d: %s",
				g.getPID(), existingMR.IID, err.Error())
			return fmt.Errorf("update GitLab merge request: %s", err.Error())
		}

		// Set auto-merge to created Merge Request
		// c.f. https://docs.gitlab.com/user/project/merge_requests/auto_merge/
		if _, _, err = g.api.MergeRequests.AcceptMergeRequest(
			g.getPID(),
			existingMR.IID,
			&acceptMergeRequestOptions,

			// client-go provides retries on rate limit (429) and server (>= 500) errors by default.
			//
			// But Method Not Allowed (405) and Unprocessable Content (422) errors will be returned
			// when AcceptMergeRequest is called immediately after CreateMergeRequest.
			//
			// c.f. https://docs.gitlab.com/api/merge_requests/#merge-a-merge-request
			//
			// Therefore, add a retryable status code only for AcceptMergeRequest calls
			gitlabapi.WithRequestRetry(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
				if ctx.Err() != nil {
					return false, ctx.Err()
				}
				if err != nil {
					return false, err
				}
				if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError || resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusUnprocessableEntity {
					return true, nil
				}
				return false, nil
			}),
		); err != nil {
			return fmt.Errorf("set auto-merge on GitLab mergerequest: %v", err)
		}

		return nil
	}

	body, err = utils.GeneratePullRequestBody("", report.ToActionsString())
	if err != nil {
		logrus.Warningf("something went wrong while generating GitLab body: %s", err)
		return fmt.Errorf("generate GitLab body: %s", err.Error())
	}

	if len(g.spec.Body) > 0 {
		body = g.spec.Body
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

	opts := gitlabapi.CreateMergeRequestOptions{
		Title:              &title,
		Description:        &body,
		SourceBranch:       &g.SourceBranch,
		TargetBranch:       &g.TargetBranch,
		AssigneeIDs:        &g.spec.Assignees,
		ReviewerIDs:        &g.spec.Reviewers,
		Squash:             g.spec.Squash,
		RemoveSourceBranch: g.spec.RemoveSourceBranch,
		Labels:             &labelOptions,
		AllowCollaboration: g.spec.AllowCollaboration,
	}

	logrus.Debugf("Title:\t%q\nBody:\t%v\nSource branch:\t%q\nTarget branch:\t%q\n",
		title,
		body,
		g.SourceBranch,
		g.TargetBranch)

	mr, resp, err := g.api.MergeRequests.CreateMergeRequest(
		g.getPID(),
		&opts,
		gitlabapi.WithContext(ctx),
	)

	if err != nil {
		logrus.Debugf("HTTP Response:\n\tReturn code: %d\n\n", resp.StatusCode)

		return fmt.Errorf("create GitLab mergerequest: %v", err)
	}

	report.Link = mr.WebURL
	report.Title = mr.Title
	report.Description = mr.Description

	// Set auto-merge to created Merge Request
	// c.f. https://docs.gitlab.com/user/project/merge_requests/auto_merge/
	if _, _, err = g.api.MergeRequests.AcceptMergeRequest(
		g.getPID(),
		mr.IID,
		&acceptMergeRequestOptions,

		// client-go provides retries on rate limit (429) and server (>= 500) errors by default.
		//
		// But Method Not Allowed (405) and Unprocessable Content (422) errors will be returned
		// when AcceptMergeRequest is called immediately after CreateMergeRequest.
		//
		// c.f. https://docs.gitlab.com/api/merge_requests/#merge-a-merge-request
		//
		// Therefore, add a retryable status code only for AcceptMergeRequest calls
		gitlabapi.WithRequestRetry(func(ctx context.Context, resp *http.Response, err error) (bool, error) {
			if ctx.Err() != nil {
				return false, ctx.Err()
			}
			if err != nil {
				return false, err
			}
			if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= http.StatusInternalServerError || resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusUnprocessableEntity {
				return true, nil
			}
			return false, nil
		}),
	); err != nil {
		return fmt.Errorf("set auto-merge on GitLab mergerequest: %v", err)
	}

	logrus.Infof("GitLab mergerequest successfully opened on %q", mr.WebURL)

	return nil
}
