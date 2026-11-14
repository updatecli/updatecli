package github

import (
	"context"
	"errors"
	"fmt"

	"github.com/shurcooL/githubv4"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/client"
)

// addComment is mutation to add a comment to a GitHub pullrequest
func (p *PullRequest) addComment(ctx context.Context, body string, retry int) error {

	if p.remotePullRequest.ID == "" {
		return nil
	}

	rateLimit, err := queryRateLimit(p.gh.client, ctx)
	if err != nil {
		if client.IsRateLimitError(err) {
			log.Debugln(rateLimit)
			if retry < client.MaxRetry {
				log.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, client.MaxRetry)
				rateLimit.Pause(ctx)
				return p.addComment(ctx, body, retry+1)
			}
			return errors.New(ErrAPIRateLimitExceededFinalAttempt)
		}
		return fmt.Errorf("unable to query GitHub API rate limit: %w", err)
	}

	log.Debugln(rateLimit)

	// cfr https://docs.github.com/en/graphql/reference/objects#issuecomment
	var mutation struct {
		AddComment struct {
			CommentEdge struct {
				Node struct {
					Body       string
					Repository struct {
						Id            string
						Name          string
						NameWithOwner string
					}
					Issue struct {
						Number int
					}
				}
			}
		} `graphql:"addComment(input: $input)"`
	}

	log.Debugf("Commenting GitHub pull request %s", p.remotePullRequest.Url)

	input := githubv4.AddCommentInput{
		SubjectID: githubv4.ID(p.remotePullRequest.ID),
		Body:      githubv4.String(body),
	}

	err = p.gh.client.Mutate(ctx, &mutation, input, nil)
	if err != nil {
		if client.IsRateLimitError(err) {
			if retry < client.MaxRetry {
				return p.addComment(ctx, body, retry+1)
			}
		}
		return fmt.Errorf("adding comment to pull request %q: %w", p.remotePullRequest.Url, err)
	}

	return nil
}
