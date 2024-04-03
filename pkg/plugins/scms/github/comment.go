package github

import (
	"context"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
)

// addComment is mutation to add a comment to a GitHub pullrequest
func (p *PullRequest) addComment(body string) error {

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
	}

	logrus.Debugf("Commenting GitHub pull request %s", p.remotePullRequest.Url)

	input := githubv4.AddCommentInput{
		SubjectID: githubv4.ID(p.remotePullRequest.ID),
		Body:      githubv4.String(body),
	}

	err := p.gh.client.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		return err
	}

	return nil
}
