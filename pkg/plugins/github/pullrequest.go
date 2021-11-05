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
# {{ .Title }}

{{ if .Introduction }}
{{ .Introduction }}
{{ end }}


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

// PullRequestSpec is a specific struct containing pullrequest settings provided
// by an updatecli configuration
type PullRequestSpec struct {
	Title                  string   // Specify pull request title
	Description            string   // Description contains user input description used during pull body creation
	Labels                 []string // Specify repository labels used for pull request. !! They must already exist
	Draft                  bool     // Define if a pull request is set to draft, default false
	MaintainerCannotModify bool     // Define if maintainer can modify pullRequest
}

func (g *Github) CreatePullRequest(title, changelog, pipelineReport string) error {

	g.pullRequest.Description = changelog
	g.pullRequest.Report = pipelineReport
	g.pullRequest.Title = title

	if len(g.spec.PullRequest.Title) > 0 {
		g.pullRequest.Title = g.spec.PullRequest.Title
	}

	err := g.getRemotePullRequest()
	if err != nil {
		return err
	}

	// No pullrequest ID so we first have to create the remote pullrequest
	if len(g.remotePullRequest.ID) == 0 {
		err = g.OpenPullRequest()
		if err != nil {
			return err
		}
	}

	// Once the remote pull request exist, we can than update it with additional information such as
	// tags,assignee,etc.

	err = g.updatePullRequest()
	if err != nil {
		return err
	}

	return nil
}

// generatePullRequestBody generates the body pull request based on PULLREQUESTBODY
func (g *Github) generatePullRequestBody() (string, error) {
	t := template.Must(template.New("pullRequest").Parse(PULLREQUESTBODY))

	buffer := new(bytes.Buffer)

	type params struct {
		Introduction string
		Title        string
		Report       string
		Description  string
	}

	err := t.Execute(buffer, params{
		Introduction: g.spec.PullRequest.Description,
		Description:  g.pullRequest.Description,
		Report:       g.pullRequest.Report,
		Title:        g.pullRequest.Title,
	})

	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// updatePullRequest updates an existing pull request.
func (g *Github) updatePullRequest() error {

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

	title := g.pullRequest.Title

	bodyPR, err := g.generatePullRequestBody()
	if err != nil {
		return err
	}

	labelsID := []githubv4.ID{}
	labels, err := g.getRepositoryLabelsInformation()
	if err != nil {
		return err
	}

	for _, label := range labels {
		labelsID = append(labelsID, githubv4.NewID(label.ID))
	}

	pullRequestUpdateStateOpen := githubv4.PullRequestUpdateStateOpen

	input := githubv4.UpdatePullRequestInput{
		PullRequestID: githubv4.String(g.remotePullRequest.ID),
		Title:         githubv4.NewString(githubv4.String(title)),
		Body:          githubv4.NewString(githubv4.String(bodyPR)),
		LabelIDs:      &labelsID,
		State:         &pullRequestUpdateStateOpen,
	}

	err = client.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		return err
	}

	logrus.Infof("\nPull Request available on:\n\n\t%s\n\n", mutation.UpdatePullRequest.PullRequest.Url)

	return nil
}

// OpenPullRequest creates a new pull request on Github.
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

	repositoryID, err := g.queryRepositoryID()
	if err != nil {
		return err
	}

	bodyPR, err := g.generatePullRequestBody()
	if err != nil {
		return err
	}

	input := githubv4.CreatePullRequestInput{
		BaseRefName:         githubv4.String(g.spec.Branch),
		RepositoryID:        githubv4.String(repositoryID),
		HeadRefName:         githubv4.String(g.remoteBranch),
		Title:               githubv4.String(g.pullRequest.Title),
		Body:                githubv4.NewString(githubv4.String(bodyPR)),
		MaintainerCanModify: githubv4.NewBoolean(githubv4.Boolean(!g.spec.PullRequest.MaintainerCannotModify)),
		Draft:               githubv4.NewBoolean(githubv4.Boolean(g.spec.PullRequest.Draft)),
	}

	err = client.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		return err
	}

	g.remotePullRequest = mutation.CreatePullRequest.PullRequest

	return nil

}

// getRemotePullRequest checks if a pull request already exist on GitHub and is in the state 'open' or 'closed'.
func (g *Github) getRemotePullRequest() error {
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
			} `graphql:"pullRequests(baseRefName: $baseRefName, headRefName: $headRefName, last: 1, states: [OPEN,CLOSED])"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":       githubv4.String(g.spec.Owner),
		"name":        githubv4.String(g.spec.Repository),
		"baseRefName": githubv4.String(g.spec.Branch),
		"headRefName": githubv4.String(g.remoteBranch),
	}

	err := client.Query(context.Background(), &query, variables)

	if err != nil {
		return err
	}

	if len(query.Repository.PullRequests.Nodes) > 0 {
		g.remotePullRequest = query.Repository.PullRequests.Nodes[0]
		return nil
	}

	return nil
}
