package pullrequest

import (
	"context"
	"fmt"
	"strings"
	"time"

	giteasdk "gitea.dev/sdk"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/result"
)

const (
	/*
		cleanupMaxAttempts defines how many times Updatecli queries the Gitea API to
		retrieve a pull request state matching the head commit of its working branch,
		before giving up cleaning that pull request.
	*/
	cleanupMaxAttempts int = 3
	// cleanupRetryDelay defines how long Updatecli waits between two attempts.
	cleanupRetryDelay time.Duration = time.Second
	// pullRequestClosedMessage explains why Updatecli closed a pull request.
	pullRequestClosedMessage string = "Pull request closed as no changed file detected"
)

var (
	// cleanupSleep is a variable so it can be overridden from the tests.
	cleanupSleep = time.Sleep
	// draftTitlePrefixes lists the title prefixes marking a pull request as a draft on a default Gitea configuration.
	draftTitlePrefixes = []string{"WIP:", "[WIP]"}
)

// draftTitlePrefix returns the prefix marking a pull request title as a draft, as written in the
// title, or an empty string if there is none.
func draftTitlePrefix(title string) string {
	for _, prefix := range draftTitlePrefixes {
		if len(title) >= len(prefix) && strings.EqualFold(title[:len(prefix)], prefix) {
			return title[:len(prefix)]
		}
	}
	return ""
}

// isDraftPullRequest tells if a pull request is held back as a draft. Gitea marks a pull request
// as a draft based on its title, and each server can configure its own prefixes.
func isDraftPullRequest(pr *giteasdk.PullRequest) bool {
	return pr.Draft || draftTitlePrefix(pr.Title) != ""
}

// existingPullRequestTitle returns the title to set on an existing pull request.
// The title follows the pipeline, but a draft must remain a draft.
func existingPullRequestTitle(title string, pr *giteasdk.PullRequest) string {
	prefix := draftTitlePrefix(pr.Title)

	switch {
	case !isDraftPullRequest(pr), draftTitlePrefix(title) != "":
		return title
	case prefix == "", title == "":
		// Without a known prefix, the server uses one we can't detect so we keep the title as it is.
		return pr.Title
	}

	return prefix + " " + title
}

// findExistingPullRequest queries a remote Gitea instance to find the open pullrequest
// matching the source and target branches. It returns nil if there is none.
// A pull request opened from a fork can use the same branch names but isn't managed
// by Updatecli, so it's ignored.
func (g *Gitea) findExistingPullRequest(ctx context.Context) (*giteasdk.PullRequest, error) {
	page := 1
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
			log.Debugf("gitea/findExistingPullRequest RC: %s\n", err)
			return nil, err
		}

		if resp != nil && resp.StatusCode > 400 {
			log.Debugf("RC: %s\nBody:\n%s", resp.Status, resp.Body)
		}

		for _, p := range pullrequests {
			// Without both branches, a pull request can't match the source and target branches.
			if p.Head == nil || p.Base == nil {
				continue
			}
			if p.Head.Name == g.SourceBranch &&
				p.Base.Name == g.TargetBranch &&
				// If no closed and merged date is set, the pullrequest is still open
				p.Closed == nil &&
				p.Merged == nil {

				if p.Head.RepoID != p.Base.RepoID {
					log.Debugf("Ignoring Gitea pull request as its head is on another repository:\n\t%s", p.HTMLURL)
					continue
				}

				log.Infof("%s GiTea pullrequest detected at:\n\t%s",
					result.SUCCESS,
					p.HTMLURL)

				return p, nil
			}
		}

		if resp == nil || resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
	}

	return nil, nil
}

/*
isPullRequestEmpty checks if the remote pull request doesn't contain any changed file.

The list endpoint never returns the number of changed files so each attempt reads the
pull request itself. That endpoint reads the head commit from the working branch while
the list endpoint reads it from the pull request ref, which Gitea updates in the
background after a push. As long as they diverge, Gitea hasn't processed the latest push
and the diff can't be trusted.

If they still diverge after maxAttempts, then we consider that we don't know enough to
close the pull request and we let a future Updatecli execution deal with it.
*/
func (g *Gitea) isPullRequestEmpty(ctx context.Context, listed *giteasdk.PullRequest, maxAttempts int) (bool, error) {
	index := listed.Index

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pullrequest, _, err := g.client.PullRequests.GetPullRequest(ctx, g.Owner, g.Repository, listed.Index)
		if err != nil {
			return false, fmt.Errorf("get Gitea pull request: %w", err)
		}

		if pullrequest.State != giteasdk.StateOpen || pullrequest.HasMerged {
			return false, nil
		}

		if pullrequest.Head != nil && listed.Head != nil &&
			pullrequest.Head.Sha != "" && pullrequest.Head.Sha == listed.Head.Sha {
			// Gitea omits the number of changed files if it failed to compute the diff
			return pullrequest.ChangedFiles != nil && *pullrequest.ChangedFiles == 0, nil
		}

		log.Debugf("Gitea pull request ref doesn't match the head commit of its working branch yet, retrying (%d/%d)",
			attempt,
			maxAttempts)

		if attempt == maxAttempts {
			break
		}

		cleanupSleep(cleanupRetryDelay)

		listed, err = g.findExistingPullRequest(ctx)
		if err != nil {
			return false, err
		}

		if listed == nil {
			// The pull request is not open anymore so there is nothing to clean
			return false, nil
		}

		if listed.Index != index {
			// Another pull request replaced the one we were asked to clean
			log.Debugf("Skipping pull request cleanup as pull request %d replaced pull request %d",
				listed.Index,
				index)
			return false, nil
		}
	}

	log.Debugf("Skipping pull request cleanup as its state doesn't reflect the working branch %q yet:\n\t%s",
		g.SourceBranch,
		listed.HTMLURL)

	return false, nil
}

// closePullRequest closes a pull request then explains why in a comment.
func (g *Gitea) closePullRequest(ctx context.Context, index int64) error {
	state := giteasdk.StateClosed

	pullrequest, _, err := g.client.PullRequests.EditPullRequest(
		ctx,
		g.Owner,
		g.Repository,
		index,
		giteasdk.EditPullRequestOption{State: &state},
	)
	if err != nil {
		return fmt.Errorf("update Gitea pull request: %w", err)
	}

	log.Infof("%s at:\n\n\t%s\n\n", pullRequestClosedMessage, pullrequest.HTMLURL)

	// Not returning an error if the comment failed to be added
	// as the main purpose of this function is to close the pullrequest
	_, _, err = g.client.Issues.CreateIssueComment(
		ctx,
		g.Owner,
		g.Repository,
		index,
		giteasdk.CreateIssueCommentOption{Body: pullRequestClosedMessage},
	)
	if err != nil {
		log.Errorf("Commenting pull-request: %s", err.Error())
	}

	return nil
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

	foundRemoteSourceBranch := false
	foundRemoteTargetBranch := false

	page := 1
	for {
		remoteBranches, resp, err := g.client.Repositories.ListRepoBranches(
			ctx,
			owner,
			repository,
			giteasdk.ListRepoBranchesOptions{
				ListOptions: giteasdk.ListOptions{
					Page:     page,
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

		if resp == nil || resp.NextPage == 0 {
			break
		}
		page = resp.NextPage
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
