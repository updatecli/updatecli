package mergerequest

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	gitlabapi "gitlab.com/gitlab-org/api/client-go"
)

// findExistingMR queries a remote GitLab instance to know if a pullrequest already exists.
func (g *Gitlab) findExistingMR() (mr *gitlabapi.BasicMergeRequest, err error) {
	ctx := context.Background()
	// Timeout api query after 30sec
	ctx, cancelList := context.WithTimeout(ctx, gitlabRequestTimeout)
	defer cancelList()

	const perPage = 30
	var page int64

	for {
		optsList := gitlabapi.ListProjectMergeRequestsOptions{
			SourceBranch: &g.SourceBranch,
			TargetBranch: &g.TargetBranch,
			State:        gitlabapi.Ptr("opened"),
			ListOptions: gitlabapi.ListOptions{
				Page:    page,
				PerPage: perPage,
			},
		}

		mergeRequests, resp, err := g.client.MergeRequests.ListProjectMergeRequests(
			g.getPID(),
			&optsList,
			gitlabapi.WithContext(ctx),
		)

		if err != nil {
			return nil, fmt.Errorf("list mrs failed with error %w", err)
		}

		page = resp.NextPage

		for _, mr := range mergeRequests {
			if mr.SourceBranch == g.SourceBranch &&
				mr.TargetBranch == g.TargetBranch &&
				mr.State == "opened" {

				logrus.Infof("%s GitLab merge request detected at:\n\t%s",
					result.SUCCESS,
					mr.WebURL)

				return mr, nil
			}
		}
		if page == 0 {
			break
		}
	}

	return nil, nil
}

// isRemoteBranchesExist queries a remote GitLab instance to know if both the pull-request source branch and the target branch exist.
func (g *Gitlab) isRemoteBranchesExist() (bool, error) {

	var sourceBranch string
	var targetBranch string
	var owner string
	var repository string

	if g.scm != nil {
		_, sourceBranch, targetBranch = g.scm.GetBranches()
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
	var page int64
	const perPage = 30
	for {
		ctx, cancel := context.WithTimeout(ctx, gitlabRequestTimeout)
		defer cancel()
		remoteBranches, resp, err := g.client.Branches.ListBranches(
			g.getPID(),
			&gitlabapi.ListBranchesOptions{
				ListOptions: gitlabapi.ListOptions{
					Page:    page,
					PerPage: perPage,
				},
			},
			gitlabapi.WithContext(ctx),
		)

		if err != nil {
			return false, fmt.Errorf("list branches failed with status code: %w", err)
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
		if page >= resp.TotalPages {
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
		_, g.SourceBranch, g.TargetBranch = g.scm.GetBranches()
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
