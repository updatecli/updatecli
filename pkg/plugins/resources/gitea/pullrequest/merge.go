package pullrequest

import (
	"context"
	"errors"
	"fmt"
	"strings"

	giteasdk "gitea.dev/sdk"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/resources/gitea/client"
)

func (g *Gitea) enableAutoMerge(ctx context.Context, pr *giteasdk.PullRequest) error {
	if !strings.EqualFold(g.spec.Merge.Strategy, MergeStrategyAuto) {
		return nil
	}

	if isDraftPullRequest(pr) {
		log.Infof("Skipping Gitea pullrequest auto-merge as %q is a draft", pr.HTMLURL)
		return nil
	}

	method := giteasdk.MergeStyle(strings.ToLower(g.spec.MergeMethod))
	if method == "" {
		repo, _, err := g.client.Repositories.GetRepo(ctx, g.Owner, g.Repository)
		if err != nil {
			return fmt.Errorf("getting repository default merge method: %w", err)
		}
		method = repo.DefaultMergeStyle
		if !isMergeMethodValid(string(method)) {
			return fmt.Errorf("unsupported repository default merge method %q", method)
		}
	}
	if pr.Head == nil || pr.Head.Sha == "" {
		return fmt.Errorf("cannot request auto-merge for Gitea pull request %q without a head commit ID", pr.HTMLURL)
	}

	accepted, resp, err := g.client.PullRequests.MergePullRequest(ctx, g.Owner, g.Repository, pr.Index, giteasdk.MergePullRequestOption{
		Style:                  method,
		MergeWhenChecksSucceed: true,
		HeadCommitId:           pr.Head.Sha,
	})
	if errors.Is(err, client.ErrAutoMergeAlreadyScheduled) {
		log.Infof("Gitea pullrequest auto-merge already enabled on %q; keeping the existing schedule", pr.HTMLURL)
		return nil
	}
	if err != nil {
		return fmt.Errorf("requesting auto-merge for Gitea pull request %q: %w", pr.HTMLURL, err)
	}
	// The SDK closes the response body and returns false, nil for HTTP rejections.
	if !accepted {
		return fmt.Errorf("auto-merge request for Gitea pull request %q rejected: HTTP %s", pr.HTMLURL, resp.Status)
	}
	log.Infof("Gitea pullrequest auto-merge requested on %q", pr.HTMLURL)
	return nil
}
