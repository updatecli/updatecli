package pullrequest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

	logrus.Debugf("Title:\t%q\nBody:\t%q\nSource:\n%q\ntarget:\t%q\n",
		title,
		body,
		b.SourceBranch,
		b.TargetBranch)

	pullRequestExists, pullRequestNumber, _, _, _, err := b.isPullRequestExist()
	if err != nil {
		return err
	}

	var responseTitle, responseBody, responseLink string
	if pullRequestExists {
		responseTitle, responseBody, responseLink, err = b.updatePullRequest(pullRequestNumber, title, body)
		if err != nil {
			return err
		}

		logrus.Infof("Bitbucket Cloud pull request successfully updated on %q", responseLink)
	} else {
		responseTitle, responseBody, responseLink, err = b.createPullRequest(title, body)
		if err != nil {
			return err
		}

		logrus.Infof("Bitbucket Cloud pull request successfully opened on %q", responseLink)
	}

	report.Link = responseLink
	report.Description = responseBody
	report.Title = responseTitle

	return nil
}

func (b *Bitbucket) createPullRequest(title, body string) (responseTitle string, responseBody string, link string, err error) {
	opts := scm.PullRequestInput{
		Title:  title,
		Body:   body,
		Source: b.SourceBranch,
		Target: b.TargetBranch,
	}

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

	b.logErrorResponse(resp)

	if err != nil {
		if err.Error() == scm.ErrNotFound.Error() {
			logrus.Infof("Bitbucket Cloud pull request not created, skipping")
			return "", "", "", nil
		}
		return "", "", "", err
	}

	return pr.Title, pr.Body, pr.Link, nil
}

func (b *Bitbucket) updatePullRequest(pullRequestNumber int, title, body string) (responseTitle string, responseBody string, link string, err error) {
	type requestInput struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Source      struct {
			Branch struct {
				Name string `json:"name"`
			} `json:"branch"`
		} `json:"source"`
		Destination struct {
			Branch struct {
				Name string `json:"name"`
			} `json:"branch"`
		} `json:"destination"`
	}

	in := new(requestInput)
	in.Title = title
	in.Description = body
	in.Source.Branch.Name = b.SourceBranch
	in.Destination.Branch.Name = b.TargetBranch

	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(in)
	if err != nil {
		return "", "", "", err
	}

	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	resp, err := b.client.Do(ctx, &scm.Request{
		Method: "PUT",
		Path:   fmt.Sprintf("2.0/repositories/%s/%s/pullrequests/%d", b.Owner, b.Repository, pullRequestNumber),
		Header: map[string][]string{
			"Content-Type": {"application/json"},
		},
		Body: buf,
	})

	b.logErrorResponse(resp)

	if err != nil {
		return "", "", "", err
	}

	type requestOutput struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Links       struct {
			HTML struct {
				Href string `json:"href"`
			} `json:"html"`
		} `json:"links"`
	}

	defer resp.Body.Close()

	pr := new(requestOutput)

	err = json.NewDecoder(resp.Body).Decode(pr)
	if err != nil {
		return "", "", "", err
	}

	return pr.Title, pr.Description, pr.Links.HTML.Href, nil
}

func (b *Bitbucket) logErrorResponse(resp *scm.Response) {
	if resp != nil {
		if resp.Status > 400 {
			logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
		}
	}
}
