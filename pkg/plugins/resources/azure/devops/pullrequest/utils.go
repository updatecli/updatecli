package pullrequest

import (
	"context"
	"time"

	"github.com/drone/go-scm/scm"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// isPullRequestExist queries a remote Azure DevOps instance to know if a pullrequest already exists.
func (a *Azure) isPullRequestExist() (bool, error) {
	ctx := context.Background()

	page := 0
	for {
		// Timeout api query after 30sec
		ctx, cancelList := context.WithTimeout(ctx, 30*time.Second)
		defer cancelList()

		optsSearch := scm.PullRequestListOptions{
			Page:   page,
			Size:   30,
			Open:   true,
			Closed: false,
		}

		pullrequests, resp, err := a.client.PullRequests.List(
			ctx,
			a.spec.RepoID,
			optsSearch,
		)

		logrus.Errorf("%v\n", err)

		if err != nil {
			logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
			return false, err
		}

		if resp.Status > 400 {
			logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
		}

		for _, p := range pullrequests {
			if p.Source == a.SourceBranch &&
				p.Target == a.TargetBranch &&
				!p.Closed &&
				!p.Merged {

				logrus.Infof("%s Nothing else to do, our pullrequest already exist on:\n\t%s",
					result.SUCCESS,
					p.Link)

				return true, nil
			}
		}

		if page >= resp.Page.Last {
			break
		}
		page++
	}

	return false, nil
}

// isRemoteBranchesExist queries a remote Azure DevOps instance to know if both the pull-request source branch and the target branch exist.
func (a *Azure) isRemoteBranchesExist() (bool, error) {

	var sourceBranch string
	var targetBranch string
	var owner string
	var project string
	var repoid string

	if a.scm != nil {
		sourceBranch = a.scm.HeadBranch
		targetBranch = a.scm.Spec.Branch
		owner = a.scm.Spec.Owner
		project = a.scm.Spec.Project
		repoid = a.scm.Spec.RepoID
	}

	if len(a.spec.SourceBranch) > 0 {
		sourceBranch = a.spec.SourceBranch
	}

	if len(a.spec.TargetBranch) > 0 {
		targetBranch = a.spec.TargetBranch
	}

	if len(a.spec.Owner) > 0 {
		owner = a.spec.Owner
	}

	if len(a.spec.Project) > 0 {
		project = a.spec.Project
	}

	if len(a.spec.RepoID) > 0 {
		project = a.spec.RepoID
	}

	// Timeout api query after 30sec
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	remoteBranches, resp, err := a.client.Git.ListBranches(
		ctx,
		repoid,
		scm.ListOptions{},
	)

	if err != nil {
		logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
		return false, err
	}

	if resp.Status > 400 {
		logrus.Debugf("RC: %d\nBody:\n%s", resp.Status, resp.Body)
	}

	foundRemoteSourceBranch := false
	foundRemoteTargetBranch := false

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

	if !foundRemoteSourceBranch {
		logrus.Debugf("Branch %q not found on remote repository %s/%s",
			sourceBranch,
			owner,
			project)
	}

	if !foundRemoteTargetBranch {
		logrus.Debugf("Branch %q not found on remote repository %s/%s",
			targetBranch,
			owner,
			project)
	}

	return false, nil
}

// inheritFromScm retrieve missing Azure DevOps settings from the azure scm object.
func (a *Azure) inheritFromScm() {

	if a.scm != nil {
		a.SourceBranch = a.scm.HeadBranch
		a.TargetBranch = a.scm.Spec.Branch
		a.spec.Owner = a.scm.Spec.Owner
		a.spec.Project = a.scm.Spec.Project
		a.spec.RepoID = a.scm.Spec.RepoID
	}

	if len(a.spec.SourceBranch) > 0 {
		a.SourceBranch = a.spec.SourceBranch
	}

	if len(a.spec.TargetBranch) > 0 {
		a.TargetBranch = a.spec.TargetBranch
	}
}
