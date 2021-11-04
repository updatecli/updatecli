package github

import (
	"bytes"
	"context"
	"text/template"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
)

// PULLREQUESTBODY is the pull request template used as pull request description
// Please note that triple backticks are concatenated with the literals, as they cannot be escaped
const PULLREQUESTBODY = `

## Report

{{ .Report }}

## Changelog

<details><summary>Click to expand</summary>

` + "```\n{{ .Description }}\n```" + `

</details>

## Remark

This pull request was automatically created using [Updatecli](https://www.updatecli.io).

Please report any issues with this tool [here](https://github.com/updatecli/updatecli/issues/new)

`

// PullRequest contains multiple fields mapped to Github V4 api
type PullRequest struct {
	BaseRefName string
	Body        string
	HeadRefName string
	ID          string
	State       string
	Title       string
	Url         string
}

// SetBody generates the body pull request based on PULLREQUESTBODY
func SetBody(changelog Changelog) (body string, err error) {
	t := template.Must(template.New("pullRequest").Parse(PULLREQUESTBODY))

	body = ""

	buffer := new(bytes.Buffer)

	err = t.Execute(buffer, changelog)

	if err != nil {
		return "", err
	}

	body = buffer.String()

	return body, err
}

// UpdatePullRequest updates an existing pull request.
func (g *Github) UpdatePullRequest(ID string) error {

	/*
				mutation($input: UpdatePullRequestInput!){
					updatePullRequest(input:$input){
						pullRequest{
							url
		          			title
		          			body
						}
					}
		    	}

				{
		  			"input": {
		  				"title":"xxx",
		    			"pullRequestId" : "yyy
					}
				}
	*/

	client := g.NewClient()

	var mutation struct {
		UpdatePullRequest struct {
			PullRequest PullRequest
		} `graphql:"updatePullRequest(input: $input)"`
	}

	logrus.Debugf("Updating Github pull request")

	title := g.pullRequestDescription.Title

	bodyPR, err := SetBody(g.pullRequestDescription)
	if err != nil {
		return err
	}

	labelsID := []githubv4.ID{}
	err = g.getRepositoryLabelsInformation()
	if err != nil {
		return err
	}

	for _, label := range g.repositoryLabels {
		labelsID = append(labelsID, githubv4.NewID(label.ID))
	}

	input := githubv4.UpdatePullRequestInput{
		PullRequestID: githubv4.String(ID),
		Title:         githubv4.NewString(githubv4.String(title)),
		Body:          githubv4.NewString(githubv4.String(bodyPR)),
		LabelIDs:      &labelsID,
	}

	err = client.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		return err
	}

	logrus.Infof("\nPull Request available on:\n\n\t%s\n\n", mutation.UpdatePullRequest.PullRequest.Url)

	return nil
}

// OpenPullRequest creates a new pull request.
func (g *Github) OpenPullRequest() error {

	/*
		mutation($input: CreatePullRequestInput!){
			createPullRequest(input:$input){
				pullRequest{
					url
				}
			}
		}
		{

		"input":{
			"baseRefName": "x" ,
			"repositoryId":"y",
			"headRefName": "z",
			"title",
			"body",
		}


	*/
	client := g.NewClient()

	var mutation struct {
		CreatePullRequest struct {
			PullRequest PullRequest
		} `graphql:"createPullRequest(input: $input)"`
	}

	logrus.Infof("Opening Pull Request")

	title := g.pullRequestDescription.Title

	repositoryID, err := g.queryRepositoryID()
	if err != nil {
		return err
	}
	maintainerCanModify := true
	draft := false

	bodyPR, err := SetBody(g.pullRequestDescription)
	if err != nil {
		return err
	}

	input := githubv4.CreatePullRequestInput{
		BaseRefName:         githubv4.String(g.spec.Branch),
		RepositoryID:        githubv4.String(repositoryID),
		HeadRefName:         githubv4.String(g.remoteBranch),
		Title:               githubv4.String(title),
		Body:                githubv4.NewString(githubv4.String(bodyPR)),
		MaintainerCanModify: githubv4.NewBoolean(githubv4.Boolean(maintainerCanModify)),
		Draft:               githubv4.NewBoolean(githubv4.Boolean(draft)),
	}

	err = client.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		return err
	}

	logrus.Infof("\nPull Request available on =>\n\n\t%s\n\n", mutation.CreatePullRequest.PullRequest.Url)

	return nil

}

// IsPullRequest checks if a pull request already exist and is in the state 'open'.
func (g *Github) IsPullRequest() (ID string, err error) {
	/*
			https://developer.github.com/v4/explorer/
		# Query
		query getPullRequests(
			$owner: String!,
			$name:String!,
			$baseRefName:String!,
			$headRefName:String!){
				repository(owner: $owner, name: $name) {
					pullRequests(baseRefName: $baseRefName, headRefName: $headRefName, last: 1) {
						nodes {
							state
							id
						}
					}
				}
			}
		}
		# Variables
		{
			"owner": "olblak",
			"name": "charts",
			"baseRefName": "master",
			"headRefName": "updatecli/HelmChart/2.4.0"
		}
	*/

	client := g.NewClient()

	var query struct {
		Repository struct {
			PullRequests struct {
				Nodes []PullRequest
			} `graphql:"pullRequests(baseRefName: $baseRefName, headRefName: $headRefName, last: 1, states: OPEN)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":       githubv4.String(g.spec.Owner),
		"name":        githubv4.String(g.spec.Repository),
		"baseRefName": githubv4.String(g.spec.Branch),
		"headRefName": githubv4.String(g.remoteBranch),
	}

	err = client.Query(context.Background(), &query, variables)

	if err != nil {
		return ID, err
	}

	if len(query.Repository.PullRequests.Nodes) > 0 {
		ID = query.Repository.PullRequests.Nodes[0].ID
	}

	return ID, err
}

// InitPullRequestDescription set internal github settings
func (g *Github) InitPullRequestDescription(title, body, report string) {
	g.pullRequestDescription.Description = body
	g.pullRequestDescription.Report = report
	g.pullRequestDescription.Title = title
}

func (g *Github) getRepositoryLabelsInformation() error {

	/*
		https://developer.github.com/v4/explorer/

			query($owner: String!, $name: String!) {
				rateLimit {
					cost
					remaining
					resetAt
				}
				repository(owner: $owner, name: $name){
					labels (last: 5, before: $before) {
						totalCount
						pageInfo {
							hasNextPage
							endCursor
						}
						edges {
							node {
								id
								name
								description
							}
							cursor
						}
					}
				}
			}

	*/

	// Early exit as no label information are needed
	if len(g.repositoryLabels) == 0 {
		return nil
	}

	client := g.NewClient()

	variables := map[string]interface{}{
		"owner":      githubv4.String(g.spec.Owner),
		"repository": githubv4.String(g.spec.Repository),
		"before":     (*githubv4.String)(nil),
	}

	var query struct {
		RateLimit  RateLimit
		Repository struct {
			Labels struct {
				TotalCount int
				PageInfo   PageInfo
				Edges      []struct {
					Cursor string
					Node   struct {
						ID          string
						Name        string
						Description string
					}
				}
			} `graphql:"labels(last: 5, before: $before)"`
		} `graphql:"repository(owner: $owner, name: $repository)"`
	}

	for {
		err := client.Query(context.Background(), &query, variables)

		if err != nil {
			logrus.Errorf("\t%s", err)
			return err
		}

		query.RateLimit.Show()

		for _, l := range g.spec.Labels {
			found := false
			for i := len(query.Repository.Labels.Edges) - 1; i >= 0; i-- {
				node := query.Repository.Labels.Edges[i]

				if l == node.Node.Name {
					found = true
					g.repositoryLabels = append(
						g.repositoryLabels,
						repositoryLabel{
							ID:          node.Node.ID,
							Name:        node.Node.Name,
							Description: node.Node.Description,
						})
					break

				}
			}
			if !found {
				logrus.Debugf("Label %q not defined on repository %s/%s, ignoring it", l, g.spec.Owner, g.spec.Repository)
			}

		}

		if !query.Repository.Labels.PageInfo.HasPreviousPage {
			break
		}

		variables["before"] = githubv4.NewString(githubv4.String(query.Repository.Labels.PageInfo.StartCursor))
	}

	logrus.Debugf("%d labels found", len(g.repositoryLabels))

	return nil
}
