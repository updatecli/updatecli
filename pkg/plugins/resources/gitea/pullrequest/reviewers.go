package pullrequest

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	giteasdk "gitea.dev/sdk"
	"github.com/updatecli/updatecli/pkg/core/log"
)

func (g *Gitea) requestReviewers(ctx context.Context, pr *giteasdk.PullRequest) error {
	if len(g.spec.Reviewers) == 0 {
		return nil
	}

	settings, _, err := g.client.Settings.GetGlobalAPISettings(ctx)
	if err != nil {
		return fmt.Errorf("listing reviews for Gitea pull request %q: getting API settings: %w", pr.HTMLURL, err)
	}
	if settings.MaxResponseItems <= 0 {
		return fmt.Errorf("listing reviews for Gitea pull request %q: invalid maximum page size %d", pr.HTMLURL, settings.MaxResponseItems)
	}
	pageSize := min(30, settings.MaxResponseItems)

	// Re-requesting a review can invalidate an existing approval.
	existingReviewers := make(map[string]bool)
	page := 1
	for {
		reviews, resp, err := g.client.PullRequests.ListPullReviews(
			ctx,
			g.Owner,
			g.Repository,
			pr.Index,
			giteasdk.ListPullReviewsOptions{
				ListOptions: giteasdk.ListOptions{
					Page:     page,
					PageSize: pageSize,
				},
			},
		)
		if err != nil {
			return fmt.Errorf("listing reviews for Gitea pull request %q: %w", pr.HTMLURL, err)
		}
		for _, review := range reviews {
			if review.Reviewer != nil {
				existingReviewers[strings.ToLower(review.Reviewer.UserName)] = true
			}
		}
		// This endpoint has no Link header, and hidden pending reviews can leave a page empty.
		count := resp.Header.Get("X-Total-Count")
		total, err := strconv.Atoi(count)
		if err != nil || total < 0 {
			return fmt.Errorf("listing reviews for Gitea pull request %q: invalid X-Total-Count header %q", pr.HTMLURL, count)
		}
		if page*pageSize >= total {
			break
		}
		page++
	}

	var reviewers []string
	for _, reviewer := range g.spec.Reviewers {
		username := strings.ToLower(reviewer)
		if !existingReviewers[username] {
			reviewers = append(reviewers, reviewer)
			existingReviewers[username] = true
		}
	}
	if len(reviewers) == 0 {
		return nil
	}

	log.Debugf("Requesting reviewers for pull request: %v", reviewers)
	_, err = g.client.PullRequests.CreateReviewRequests(
		ctx,
		g.Owner,
		g.Repository,
		pr.Index,
		giteasdk.PullReviewRequestOptions{
			Reviewers: reviewers,
		},
	)
	if err != nil {
		return fmt.Errorf("requesting reviewers for Gitea pull request %q: %w", pr.HTMLURL, err)
	}
	return nil
}
