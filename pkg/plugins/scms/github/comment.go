package github

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github/client"
)

// addComment is mutation to add a comment to a GitHub pullrequest
func (p *PullRequest) addComment(body string, retry int) error {

	if p.remotePullRequest.ID == "" {
		return nil
	}

	rateLimit, err := queryRateLimit(p.gh.client, context.Background())
	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			logrus.Debugln(rateLimit)
			if retry < client.MaxRetry {
				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, client.MaxRetry)
				rateLimit.Pause()
				return p.addComment(body, retry+1)
			}
			return errors.New(ErrAPIRateLimitExceededFinalAttempt)
		}
		return fmt.Errorf("unable to query GitHub API rate limit: %w", err)
	}

	logrus.Debugln(rateLimit)

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

	logrus.Debugf("Commenting GitHub pull request %s", p.remotePullRequest.Url)

	input := githubv4.AddCommentInput{
		SubjectID: githubv4.ID(p.remotePullRequest.ID),
		Body:      githubv4.String(body),
	}

	err = p.gh.client.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			if retry < client.MaxRetry {
				return p.addComment(body, retry+1)
			}
		}
		return fmt.Errorf("adding comment to pull request %q: %w", p.remotePullRequest.Url, err)
	}

	return nil
}
