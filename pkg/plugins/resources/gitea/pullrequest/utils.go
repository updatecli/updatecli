package pullrequest

import (
	"context"

	giteasdk "gitea.dev/sdk"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// isPullRequestExist queries a remote Gitea instance to know if a pullrequest already exists.
func (g *Gitea) isPullRequestExist(ctx context.Context) (*giteasdk.PullRequest, error) {
	page := 0
	for {
		optsSearch := giteasdk.ListPullRequestsOptions{
			State: "open",
			ListOptions: giteasdk.ListOptions{
				Page:     page,
				PageSize: 30,
			},
		}

		pullrequests, resp, err := g.client.PullRequests.ListRepoPullRequests(
			ctx,
			g.Owner,
			g.Repository,
			optsSearch,
		)

		if err != nil {
			log.Debugf("gitea/isPullRequestExist RC: %s\n", err)
			return nil, err
		}

		if resp != nil && resp.StatusCode > 400 {
			log.Debugf("RC: %s\nBody:\n%s", resp.Status, resp.Body)
		}

		for _, p := range pullrequests {
			// A fork can use the same branch name as the working branch.
			if p.Head == nil || p.Base == nil || p.Head.RepoID != p.Base.RepoID {
				continue
			}
			if p.Head.Name == g.SourceBranch &&
				p.Base.Name == g.TargetBranch &&
				// If no closed and merged date is set, the pullrequest is still open
				p.Closed == nil &&
				p.Merged == nil {

				log.Infof("%s GiTea pullrequest detected at:\n\t%s",
					result.SUCCESS,
					p.URL)

				return p, nil
			}
		}

		if page >= resp.LastPage {
			break
		}
		page++
	}

	return nil, nil
}

// isRemoteBranchesExist queries a remote Gitea instance to know if both the pull-request source branch and the target branch exist.
func (g *Gitea) isRemoteBranchesExist(ctx context.Context) (bool, error) {

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

	remoteBranches, resp, err := g.client.Repositories.ListRepoBranches(
		ctx,
		owner,
		repository,
		giteasdk.ListRepoBranchesOptions{
			ListOptions: giteasdk.ListOptions{
				Page:     1,
				PageSize: 30,
			},
		},
	)

	if err != nil {
		log.Debugf("Error: %s", err)
		return false, err
	}

	if resp != nil && resp.StatusCode > 400 {
		log.Debugf("gitea/ListRepoBranches RC: %d", resp.StatusCode)
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
		log.Debugf("Branch %q not found on remote repository %s/%s",
			sourceBranch,
			owner,
			repository)
	}

	if !foundRemoteTargetBranch {
		log.Debugf("Branch %q not found on remote repository %s/%s",
			targetBranch,
			owner,
			repository)
	}

	return false, nil
}

// inheritFromScm retrieve missing gitea settings from the gitea scm object.
func (g *Gitea) inheritFromScm() {

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
