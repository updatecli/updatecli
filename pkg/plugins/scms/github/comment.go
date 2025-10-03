package github

import (
	"context"
	"fmt"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
)

// addComment is mutation to add a comment to a GitHub pullrequest
func (p *PullRequest) addComment(body string, retry int) error {

	if p.remotePullRequest.ID == "" {
		return nil
	}

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
		RateLimit RateLimit
	}

	logrus.Debugf("Commenting GitHub pull request %s", p.remotePullRequest.Url)

	input := githubv4.AddCommentInput{
		SubjectID: githubv4.ID(p.remotePullRequest.ID),
		Body:      githubv4.String(body),
	}

	err := p.gh.client.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			logrus.Debugln(mutation.RateLimit)
			if retry < MaxRetry {
				mutation.RateLimit.Pause()

				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, MaxRetry)
				return p.addComment(body, retry+1)
			}
			return fmt.Errorf("%s", ErrAPIRateLimitExceededFinalAttempt)
		}
		return fmt.Errorf("adding comment to pull request %q: %w", p.remotePullRequest.Url, err)
	}

	logrus.Debugln(mutation.RateLimit)

	return nil
}
