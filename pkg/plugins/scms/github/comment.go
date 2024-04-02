package github

import (
	"context"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
)

// IssueComment represents the data returned
// by the GitHub API when adding a comment to an issue
// cfr https://docs.github.com/en/graphql/reference/objects#issuecomment
type IssueComment struct {
	ID   string
	URL  string
	Body string
}

// addComment is mutation to add a comment to a GitHub pullrequest
func (p *PullRequest) addComment(ID, body string) error {

	var mutation struct {
		AddComment struct {
			Comment IssueComment
		} `graphql:"addComment(input: $input)"`
	}

	logrus.Debugf("Commenting GitHub pull-request %s", ID)

	input := githubv4.AddCommentInput{
		SubjectID: githubv4.ID(ID),
		Body:      githubv4.String(body),
	}

	err := p.gh.client.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		return err
	}

	return nil
}
