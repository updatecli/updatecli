package github

/*
	Keeping the pullrequest code under the github package to avoid cyclic dependencies between github <-> pullrequest
	We need to completely refactor the github package to split the different component into specific sub packages

		github/pullrequest
		github/target
		github/scm
		github/source
		github/condition
*/

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
type PullRequestApi struct {
	BaseRefName string
	Body        string
	HeadRefName string
	ID          string
	State       string
	Title       string
	Url         string
	Number      int
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

type PullRequest struct {
	gh                *Github
	Description       string
	Report            string
	Title             string
	spec              PullRequestSpec
	remotePullRequest PullRequestApi
}

func NewPullRequest(spec PullRequestSpec, gh *Github) (PullRequest, error) {
	return PullRequest{
		gh:   gh,
		spec: spec,
	}, nil
}

func (p *PullRequest) CreatePullRequest(title, changelog, pipelineReport string) error {

	p.Description = changelog
	p.Report = pipelineReport
	p.Title = title

	// Check if there is already a pullRequest for current pipeline
	err := p.getRemotePullRequest()
	if err != nil {
		return err
	}

	// No pullrequest ID so we first have to create the remote pullrequest
	if len(p.remotePullRequest.ID) == 0 {
		err = p.OpenPullRequest()
		if err != nil {
			return err
		}
	}

	// Once the remote pull request exist, we can than update it with additional information such as
	// tags,assignee,etc.

	err = p.updatePullRequest()
	if err != nil {
		return err
	}

	return nil
}

// generatePullRequestBody generates the body pull request based on PULLREQUESTBODY
func (p *PullRequest) generatePullRequestBody() (string, error) {
	t := template.Must(template.New("pullRequest").Parse(PULLREQUESTBODY))

	buffer := new(bytes.Buffer)

	type params struct {
		Introduction string
		Title        string
		Report       string
		Description  string
	}

	err := t.Execute(buffer, params{
		Introduction: p.spec.Description,
		Description:  p.Description,
		Report:       p.Report,
		Title:        p.Title,
	})

	if err != nil {
		return "", err
	}

	return buffer.String(), nil
}

// updatePullRequest updates an existing pull request.
func (p *PullRequest) updatePullRequest() error {

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

	client := p.gh.NewClient()

	var mutation struct {
		UpdatePullRequest struct {
			PullRequest PullRequestApi
		} `graphql:"updatePullRequest(input: $input)"`
	}

	logrus.Debugf("Updating Github pull request")

	title := p.Title

	bodyPR, err := p.generatePullRequestBody()
	if err != nil {
		return err
	}

	labelsID := []githubv4.ID{}
	repolabels, err := p.gh.GetRepositoryLabelsInformation()
	if err != nil {
		return err
	}

	remotePRLabels, err := p.GetPullRequestLabelsInformation()
	if err != nil {
		return err
	}

	// Build a list of labelID to update the pullrequest
	for _, label := range MergeLabels(repolabels, remotePRLabels) {
		labelsID = append(labelsID, githubv4.NewID(label.ID))
	}

	pullRequestUpdateStateOpen := githubv4.PullRequestUpdateStateOpen

	input := githubv4.UpdatePullRequestInput{
		PullRequestID: githubv4.String(p.remotePullRequest.ID),
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
func (p *PullRequest) OpenPullRequest() error {

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
	client := p.gh.NewClient()

	var mutation struct {
		CreatePullRequest struct {
			PullRequest PullRequestApi
		} `graphql:"createPullRequest(input: $input)"`
	}

	repositoryID, err := p.gh.queryRepositoryID()
	if err != nil {
		return err
	}

	bodyPR, err := p.generatePullRequestBody()
	if err != nil {
		return err
	}

	input := githubv4.CreatePullRequestInput{
		BaseRefName:         githubv4.String(p.gh.Spec.Branch),
		RepositoryID:        githubv4.String(repositoryID),
		HeadRefName:         githubv4.String(p.gh.HeadBranch),
		Title:               githubv4.String(p.Title),
		Body:                githubv4.NewString(githubv4.String(bodyPR)),
		MaintainerCanModify: githubv4.NewBoolean(githubv4.Boolean(!p.spec.MaintainerCannotModify)),
		Draft:               githubv4.NewBoolean(githubv4.Boolean(p.spec.Draft)),
	}

	err = client.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		return err
	}

	p.remotePullRequest = mutation.CreatePullRequest.PullRequest

	return nil

}

// getRemotePullRequest checks if a pull request already exist on GitHub and is in the state 'open' or 'closed'.
func (p *PullRequest) getRemotePullRequest() error {
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

	client := p.gh.NewClient()

	var query struct {
		Repository struct {
			PullRequests struct {
				Nodes []PullRequestApi
			} `graphql:"pullRequests(baseRefName: $baseRefName, headRefName: $headRefName, last: 1, states: [OPEN,CLOSED])"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":       githubv4.String(p.gh.Spec.Owner),
		"name":        githubv4.String(p.gh.Spec.Repository),
		"baseRefName": githubv4.String(p.gh.Spec.Branch),
		"headRefName": githubv4.String(p.gh.HeadBranch),
	}

	err := client.Query(context.Background(), &query, variables)

	if err != nil {
		return err
	}

	if len(query.Repository.PullRequests.Nodes) > 0 {
		p.remotePullRequest = query.Repository.PullRequests.Nodes[0]
		return nil
	}

	return nil
}

// getPullRequestLabelsInformation queries GitHub Api to retrieve every labels assigned to a pullRequest
func (p *PullRequest) GetPullRequestLabelsInformation() ([]repositoryLabelApi, error) {

	/*
		query getPullRequests(
			$owner: String!,
			$name:String!,
			$before:Int!){
				repository(owner: $owner, name: $name) {
					pullRequest(number: 4){
		            labels(last: 5, before:$before){
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
					}
	*/

	client := p.gh.NewClient()

	variables := map[string]interface{}{
		"owner":      githubv4.String(p.gh.Spec.Owner),
		"repository": githubv4.String(p.gh.Spec.Repository),
		"number":     githubv4.Int(p.remotePullRequest.Number),
		"before":     (*githubv4.String)(nil),
	}

	var query struct {
		RateLimit  RateLimit
		Repository struct {
			PullRequest struct {
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
			} `graphql:"pullRequest(number: $number)"`
		} `graphql:"repository(owner: $owner, name: $repository)"`
	}

	var pullRequestLabels []repositoryLabelApi
	for {
		err := client.Query(context.Background(), &query, variables)

		if err != nil {
			logrus.Errorf("\t%s", err)
			return nil, err
		}

		query.RateLimit.Show()

		// Retrieve remote label information such as label ID, label name, labe description
		for _, node := range query.Repository.PullRequest.Labels.Edges {
			pullRequestLabels = append(
				pullRequestLabels,
				repositoryLabelApi{
					ID:          node.Node.ID,
					Name:        node.Node.Name,
					Description: node.Node.Description,
				})
		}

		if !query.Repository.PullRequest.Labels.PageInfo.HasPreviousPage {
			break
		}

		variables["before"] = githubv4.NewString(githubv4.String(query.Repository.PullRequest.Labels.PageInfo.StartCursor))
	}
	return pullRequestLabels, nil
}
