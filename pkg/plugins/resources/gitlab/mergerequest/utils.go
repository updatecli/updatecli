package mergerequest

import (
	"context"
	"strings"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// isPullRequestExist queries a remote GitLab instance to know if a pullrequest already exists.
func (g *Gitlab) isPullRequestExist() (bool, error) {
	ctx := context.Background()
	// Timeout api query after 30sec
	ctx, cancelList := context.WithTimeout(ctx, 30*time.Second)
	defer cancelList()

	page := 0
	for {
		optsSearch := scm.PullRequestListOptions{
			Page:   page,
			Size:   30,
			Open:   true,
			Closed: false,
		}

		pullrequests, resp, err := g.client.PullRequests.List(
			ctx,
			strings.Join([]string{
				g.Owner,
				g.Repository}, "/"),
			optsSearch,
		)

		if err != nil {
			logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
			return false, err
		}

		page = resp.Page.Next

		if resp.Status > 400 {
			logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
		}

		for _, p := range pullrequests {
			if p.Source == g.SourceBranch &&
				p.Target == g.TargetBranch &&
				!p.Closed &&
				!p.Merged {

				logrus.Infof("%s Nothing else to do, our pullrequest already exist on:\n\t%s",
					result.SUCCESS,
					p.Link)

				return true, nil
			}
		}
		if page == 0 {
			break
		}
	}

	return false, nil
}

// isRemoteBranchesExist queries a remote GitLab instance to know if both the pull-request source branch and the target branch exist.
func (g *Gitlab) isRemoteBranchesExist() (bool, error) {

	var sourceBranch string
	var targetBranch string
	var owner string
	var repository string

	if g.scm != nil {
		sourceBranch = g.scm.HeadBranch
		targetBranch = g.scm.Spec.Branch
		owner = g.scm.Spec.Owner
		repository = g.scm.Spec.Repository
	}

	if len(g.spec.SourceBranch) > 0 {
		sourceBranch = g.spec.SourceBranch
	}

	if len(g.spec.TargetBranch) > 0 {
		targetBranch = g.spec.TargetBranch
	}

	if len(g.spec.Owner) > 0 {
		owner = g.spec.Owner
	}

	if len(g.spec.Repository) > 0 {
		repository = g.spec.Repository
	}

	// Timeout api query after 30sec
	ctx := context.Background()

	foundRemoteSourceBranch := false
	foundRemoteTargetBranch := false
	page := 0
	for {
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		remoteBranches, resp, err := g.client.Git.ListBranches(
			ctx,
			strings.Join([]string{owner, repository}, "/"),
			scm.ListOptions{
				URL:  g.spec.URL,
				Page: page,
				Size: 30,
			},
		)

		if err != nil {
			logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
			return false, err
		}

		if resp.Status > 400 {
			logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
		}

		for _, remoteBranch := range remoteBranches {
			if remoteBranch.Name == sourceBranch {
				foundRemoteSourceBranch = true
			}
			if remoteBranch.Name == targetBranch {
				foundRemoteTargetBranch = true
			}

			if foundRemoteSourceBranch && foundRemoteTargetBranch {
				return true, nil
			}
		}
		if page >= resp.Page.Last {
			break
		}
		page++
	}

	if !foundRemoteSourceBranch {
		logrus.Debugf("Branch %q not found on remote repository %s/%s",
			sourceBranch,
			owner,
			repository)
	}

	if !foundRemoteTargetBranch {
		logrus.Debugf("Branch %q not found on remote repository %s/%s",
			targetBranch,
			owner,
			repository)
	}

	return false, nil
}

// inheritFromScm retrieve missing GitLab settings from the GitLab scm object.
func (g *Gitlab) inheritFromScm() {

	if g.scm != nil {
		g.SourceBranch = g.scm.HeadBranch
		g.TargetBranch = g.scm.Spec.Branch
		g.Owner = g.scm.Spec.Owner
		g.Repository = g.scm.Spec.Repository
	}

	if len(g.spec.SourceBranch) > 0 {
		g.SourceBranch = g.spec.SourceBranch
	}

	if len(g.spec.TargetBranch) > 0 {
		g.TargetBranch = g.spec.TargetBranch
	}

	if len(g.spec.Owner) > 0 {
		g.Owner = g.spec.Owner
	}

	if len(g.spec.Repository) > 0 {
		g.Repository = g.spec.Repository
	}
}
