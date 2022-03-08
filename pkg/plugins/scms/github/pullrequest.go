package github

import (
	"context"
	"errors"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
)

var ErrAutomergeNotAllowOnRepository = errors.New("automerge is not allowed on repository")

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

// Graphql mutation used with GitHub api to enable automerge on a existing
// pullrequest
type mutationEnablePullRequestAutoMerge struct {
	EnablePullRequestAutoMerge struct {
		PullRequest PullRequestApi
	} `graphql:"enablePullRequestAutoMerge(input: $input)"`
}

// updatePullRequest updates an existing pull request.
func (gh *Github) UpdatePullRequest(newTitle, bodyPR string, labels []string, remotePullRequest PullRequestApi) error {
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
	var mutation struct {
		UpdatePullRequest struct {
			PullRequest PullRequestApi
		} `graphql:"updatePullRequest(input: $input)"`
	}

	logrus.Debugf("Updating Github pull request")

	labelsID := []githubv4.ID{}
	repositoryLabels, err := gh.GetRepositoryLabels()
	if err != nil {
		return err
	}

	matchingLabels := []RepositoryLabelApi{}
	for _, l := range labels {
		for _, repoLabel := range repositoryLabels {
			if l == repoLabel.Name {
				matchingLabels = append(matchingLabels, repoLabel)
			}
		}
	}

	remotePRLabels, err := gh.getPullRequestLabelsInformation(remotePullRequest)
	if err != nil {
		return err
	}

	// Build a list of labelID to update the pullrequest
	for _, label := range MergeLabels(matchingLabels, remotePRLabels) {
		labelsID = append(labelsID, githubv4.NewID(label.ID))
	}

	pullRequestUpdateStateOpen := githubv4.PullRequestUpdateStateOpen

	input := githubv4.UpdatePullRequestInput{
		PullRequestID: githubv4.String(remotePullRequest.ID),
		Title:         githubv4.NewString(githubv4.String(remotePullRequest.Title)),
		Body:          githubv4.NewString(githubv4.String(bodyPR)),
		LabelIDs:      &labelsID,
		State:         &pullRequestUpdateStateOpen,
	}

	err = gh.client.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		return err
	}

	logrus.Infof("\nPull Request available at:\n\n\t%s\n\n", mutation.UpdatePullRequest.PullRequest.Url)

	return nil
}

// OpenPullRequest creates a new pull request on Github.
func (gh *Github) OpenPullRequest(title, bodyPR string, maintainerCannotModify, isDraft bool) (PullRequestApi, error) {
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
		}
	*/
	var mutation struct {
		CreatePullRequest struct {
			PullRequest PullRequestApi
		} `graphql:"createPullRequest(input: $input)"`
	}

	repositoryID, err := gh.queryRepositoryID()
	if err != nil {
		return PullRequestApi{}, err
	}

	input := githubv4.CreatePullRequestInput{
		BaseRefName:         githubv4.String(gh.Spec.Branch),
		RepositoryID:        githubv4.String(repositoryID),
		HeadRefName:         githubv4.String(gh.HeadBranch),
		Title:               githubv4.String(title),
		Body:                githubv4.NewString(githubv4.String(bodyPR)),
		MaintainerCanModify: githubv4.NewBoolean(githubv4.Boolean(!maintainerCannotModify)),
		Draft:               githubv4.NewBoolean(githubv4.Boolean(isDraft)),
	}

	err = gh.client.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		return PullRequestApi{}, err
	}

	return mutation.CreatePullRequest.PullRequest, nil
}

// EnablePullRequestAutoMerge updates an existing pullrequest with the flag automerge
func (gh *Github) EnablePullRequestAutoMerge(mergeMethod string, remotePullRequest PullRequestApi) error {
	// Test that automerge feature is enabled on repository but only if we plan to use it
	autoMergeAllowed, err := gh.isAutoMergedEnabledOnRepository()
	if err != nil {
		return err
	}

	if !autoMergeAllowed {
		return ErrAutomergeNotAllowOnRepository
	}

	var mutation mutationEnablePullRequestAutoMerge

	input := githubv4.EnablePullRequestAutoMergeInput{
		PullRequestID: githubv4.String(remotePullRequest.ID),
	}

	// The Github Api expects the merge method to be capital letter and don't allows empty value
	// hence the reason to set input.MergeMethod only if the value is not nil
	if len(mergeMethod) > 0 {
		mergeMethod := githubv4.PullRequestMergeMethod(strings.ToUpper(mergeMethod))
		input.MergeMethod = &mergeMethod
	}

	err = gh.client.Mutate(context.Background(), &mutation, input, nil)

	if err != nil {
		return err
	}

	return nil
}

// isAutoMergedEnabledOnRepository checks if a remote repository allows automerging pull requests
func (gh *Github) isAutoMergedEnabledOnRepository() (bool, error) {
	var query struct {
		Repository struct {
			AutoMergeAllowed bool
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner": githubv4.String(gh.Spec.Owner),
		"name":  githubv4.String(gh.Spec.Repository),
	}

	err := gh.client.Query(context.Background(), &query, variables)

	if err != nil {
		return false, err
	}
	return query.Repository.AutoMergeAllowed, nil
}

// getPullRequestLabelsInformation queries GitHub Api to retrieve every labels assigned to a pullRequest
func (gh *Github) getPullRequestLabelsInformation(remotePullRequest PullRequestApi) ([]RepositoryLabelApi, error) {
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

	variables := map[string]interface{}{
		"owner":      githubv4.String(gh.Spec.Owner),
		"repository": githubv4.String(gh.Spec.Repository),
		"number":     githubv4.Int(remotePullRequest.Number),
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

	var pullRequestLabels []RepositoryLabelApi
	for {
		err := gh.client.Query(context.Background(), &query, variables)

		if err != nil {
			logrus.Errorf("\t%s", err)
			return nil, err
		}

		query.RateLimit.Show()

		// Retrieve remote label information such as label ID, label name, labe description
		for _, node := range query.Repository.PullRequest.Labels.Edges {
			pullRequestLabels = append(
				pullRequestLabels,
				RepositoryLabelApi{
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

// getRemotePullRequest checks if a pull request already exist on GitHub and is in the state 'open' or 'closed'.
func (gh *Github) GetRemotePullRequest() (PullRequestApi, error) {
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

	var query struct {
		Repository struct {
			PullRequests struct {
				Nodes []PullRequestApi
			} `graphql:"pullRequests(baseRefName: $baseRefName, headRefName: $headRefName, last: 1, states: [OPEN])"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":       githubv4.String(gh.Spec.Owner),
		"name":        githubv4.String(gh.Spec.Repository),
		"baseRefName": githubv4.String(gh.Spec.Branch),
		"headRefName": githubv4.String(gh.HeadBranch),
	}

	err := gh.client.Query(context.Background(), &query, variables)

	if err != nil {
		return PullRequestApi{}, err
	}

	if len(query.Repository.PullRequests.Nodes) > 0 {
		return query.Repository.PullRequests.Nodes[0], nil
	}

	return PullRequestApi{}, nil
}
