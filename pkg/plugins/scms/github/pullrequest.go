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
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"

	"github.com/updatecli/updatecli/pkg/core/reports"
	utils "github.com/updatecli/updatecli/pkg/plugins/utils/action"
)

var (
	ErrAutomergeNotAllowOnRepository = errors.New("automerge is not allowed on repository")
	ErrBadMergeMethod                = errors.New("wrong merge method defined, accepting one of 'squash', 'merge', 'rebase', or ''")
)

// PullRequest contains multiple fields mapped to GitHub V4 api
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

// ActionSpec specifies the configuration of an action of type "GitHub Pull Request"
type ActionSpec struct {
	// Specifies if automerge is enabled for the new pullrequest
	AutoMerge bool `yaml:",omitempty"`
	// Specifies the Pull Request title
	Title string `yaml:",omitempty"`
	// Specifies user input description used during pull body creation
	Description string `yaml:",omitempty"`
	// Specifies repository labels used for the Pull Request. !! Labels must already exist on the repository
	Labels []string `yaml:",omitempty"`
	// Specifies if a Pull Request is set to draft, default false
	Draft bool `yaml:",omitempty"`
	// Specifies if maintainer can modify pullRequest
	MaintainerCannotModify bool `yaml:",omitempty"`
	// Specifies which merge method is used to incorporate the Pull Request. Accept "merge", "squash", "rebase", or ""
	MergeMethod string `yaml:",omitempty"`
	// Specifies to use the Pull Request title as commit message when using auto merge, only works for "squash" or "rebase"
	UseTitleForAutoMerge bool `yaml:",omitempty"`
	// Specifies if a Pull Request should be sent to the parent of a fork.
	Parent bool `yaml:",omitempty"`
}

type PullRequest struct {
	gh                *Github
	Report            string
	Title             string
	spec              ActionSpec
	remotePullRequest PullRequestApi
	repository        *Repository
}

// isMergeMethodValid ensure that we specified a valid merge method.
func isMergeMethodValid(method string) (bool, error) {
	if len(method) == 0 ||
		strings.ToUpper(method) == "SQUASH" ||
		strings.ToUpper(method) == "MERGE" ||
		strings.ToUpper(method) == "REBASE" {
		return true, nil
	}
	logrus.Debugf("%s - %s", method, ErrBadMergeMethod)
	return false, ErrBadMergeMethod
}

// Validate ensures that the provided ActionSpec is valid
func (s *ActionSpec) Validate() error {

	if _, err := isMergeMethodValid(s.MergeMethod); err != nil {
		return err
	}
	return nil
}

// Graphql mutation used with GitHub api to enable automerge on a existing
// pullrequest
type mutationEnablePullRequestAutoMerge struct {
	EnablePullRequestAutoMerge struct {
		PullRequest PullRequestApi
	} `graphql:"enablePullRequestAutoMerge(input: $input)"`
}

func NewAction(spec ActionSpec, gh *Github) (PullRequest, error) {
	err := spec.Validate()

	return PullRequest{
		gh:   gh,
		spec: spec,
	}, err
}

func (p *PullRequest) CreateAction(report reports.Action) error {

	// One GitHub pullrequest body can contain multiple action report
	// It would be better to refactor CreateAction
	p.Report = report.ToActionsString()
	p.Title = report.Title

	if p.spec.Title != "" {
		p.Title = p.spec.Title
	}

	repository, err := p.gh.queryRepository()
	if err != nil {
		return err
	}

	p.repository = repository

	// Check if they are changes that need to be published otherwise exit
	matchingBranch, err := p.gh.nativeGitHandler.IsSimilarBranch(
		p.gh.HeadBranch,
		p.gh.Spec.Branch,
		p.gh.GetDirectory())

	if err != nil {
		return err
	}

	if matchingBranch {
		logrus.Debugf("GitHub pullrequest not needed")

		return nil
	}

	// Check if there is already a pullRequest for current pipeline
	err = p.getRemotePullRequest()
	if err != nil {
		return err
	}

	// If there is not already a Pull Request ID then no existing PR so let's open it
	if len(p.remotePullRequest.ID) == 0 {
		err = p.OpenPullRequest()
		if err != nil {
			return err
		}
	}

	// Once the remote Pull Request exists, we can than update it with additional information such as
	// tags,assignee,etc.

	err = p.updatePullRequest()
	if err != nil {
		return err
	}

	if p.spec.AutoMerge {
		err = p.EnablePullRequestAutoMerge()
		if err != nil {
			return err
		}
	}

	return nil
}

// updatePullRequest updates an existing Pull Request.
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
			  "pullRequestId" : "yyy"
			}
		  }
	*/
	var mutation struct {
		UpdatePullRequest struct {
			PullRequest PullRequestApi
		} `graphql:"updatePullRequest(input: $input)"`
	}

	logrus.Debugf("Updating GitHub Pull Request")

	title := p.Title

	bodyPR, err := utils.GeneratePullRequestBody(p.spec.Description, p.Report)
	if err != nil {
		return err
	}

	labelsID := []githubv4.ID{}
	repositoryLabels, err := p.gh.getRepositoryLabels()
	if err != nil {
		logrus.Debugf("Error fetching repository labels: %s", err.Error())
		return err
	}

	matchingLabels := []repositoryLabelApi{}
	for _, l := range p.spec.Labels {
		for _, repoLabel := range repositoryLabels {
			if l == repoLabel.Name {
				matchingLabels = append(matchingLabels, repoLabel)
			}
		}
	}

	remotePRLabels, err := p.GetPullRequestLabelsInformation()
	if err != nil {
		logrus.Debugf("Error fetching labels information: %s", err.Error())
		return err
	}

	// Build a list of labelID to update the pullrequest
	for _, label := range mergeLabels(matchingLabels, remotePRLabels) {
		labelsID = append(labelsID, githubv4.NewID(label.ID))
	}

	input := githubv4.UpdatePullRequestInput{
		PullRequestID: githubv4.ID(p.remotePullRequest.ID),
		Title:         githubv4.NewString(githubv4.String(title)),
		Body:          githubv4.NewString(githubv4.String(bodyPR)),
	}

	if len(p.spec.Labels) != 0 {
		input.LabelIDs = &labelsID
	}

	err = p.gh.client.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		logrus.Debugf("Error updating pull-request: %s", err.Error())
		return err
	}

	logrus.Infof("\nPull Request available at:\n\n\t%s\n\n", mutation.UpdatePullRequest.PullRequest.Url)

	return nil
}

// EnablePullRequestAutoMerge updates an existing pullrequest with the flag automerge
func (p *PullRequest) EnablePullRequestAutoMerge() error {

	// Test that automerge feature is enabled on repository but only if we plan to use it
	autoMergeAllowed, err := p.isAutoMergedEnabledOnRepository()
	if err != nil {
		return err
	}

	if !autoMergeAllowed {
		return ErrAutomergeNotAllowOnRepository
	}

	var mutation mutationEnablePullRequestAutoMerge

	input := githubv4.EnablePullRequestAutoMergeInput{
		PullRequestID: githubv4.String(p.remotePullRequest.ID),
	}

	// The GitHub Api expects the merge method to be capital letter and don't allows empty value
	// hence the reason to set input.MergeMethod only if the value is not nil
	if len(p.spec.MergeMethod) > 0 {
		mergeMethod := githubv4.PullRequestMergeMethod(strings.ToUpper(p.spec.MergeMethod))
		input.MergeMethod = &mergeMethod
	}

	if p.spec.UseTitleForAutoMerge {
		if strings.EqualFold(p.spec.MergeMethod, "squash") {
			input.CommitHeadline = githubv4.NewString(githubv4.String(fmt.Sprintf("%s (#%d)", p.Title, p.remotePullRequest.Number)))
		} else if strings.EqualFold(p.spec.MergeMethod, "rebase") {
			input.CommitHeadline = githubv4.NewString(githubv4.String(p.Title))
		}
	}

	err = p.gh.client.Mutate(context.Background(), &mutation, input, nil)

	if err != nil {
		return err
	}

	return nil
}

// OpenPullRequest creates a new GitHub Pull Request.
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
	       "headRepositoryId": "",
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

	bodyPR, err := utils.GeneratePullRequestBody(p.spec.Description, p.Report)
	if err != nil {
		return err
	}

	input := githubv4.CreatePullRequestInput{
		RepositoryID:        githubv4.String(p.repository.ID),
		BaseRefName:         githubv4.String(p.gh.Spec.Branch),
		HeadRefName:         githubv4.String(p.gh.HeadBranch),
		Title:               githubv4.String(p.Title),
		Body:                githubv4.NewString(githubv4.String(bodyPR)),
		MaintainerCanModify: githubv4.NewBoolean(githubv4.Boolean(!p.spec.MaintainerCannotModify)),
		Draft:               githubv4.NewBoolean(githubv4.Boolean(p.spec.Draft)),
	}

	if p.spec.Parent {
		input.RepositoryID = githubv4.String(p.repository.ParentID)
		input.HeadRepositoryID = githubv4.NewID(p.repository.ID)
	}

	logrus.Debugf("Opening pull-request against repo %s", input.RepositoryID)

	err = p.gh.client.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		logrus.Infof("\nError creating pull request:\n\n\t%s\n\n", err.Error())
		return err
	}

	p.remotePullRequest = mutation.CreatePullRequest.PullRequest

	return nil

}

// isAutoMergedEnabledOnRepository checks if a remote repository allows automerging Pull Requests.
func (p *PullRequest) isAutoMergedEnabledOnRepository() (bool, error) {

	var query struct {
		Repository struct {
			AutoMergeAllowed bool
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner": githubv4.String(p.gh.Spec.Owner),
		"name":  githubv4.String(p.gh.Spec.Repository),
	}

	err := p.gh.client.Query(context.Background(), &query, variables)

	if err != nil {
		return false, err
	}
	return query.Repository.AutoMergeAllowed, nil

}

// getRemotePullRequest checks if a Pull Request already exists on GitHub and is in the state 'open' or 'closed'.
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

	var query struct {
		Repository struct {
			PullRequests struct {
				Nodes []PullRequestApi
			} `graphql:"pullRequests(baseRefName: $baseRefName, headRefName: $headRefName, last: 1, states: [OPEN])"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	owner := githubv4.String(p.repository.Owner)
	name := githubv4.String(p.repository.Name)

	if p.spec.Parent {
		owner = githubv4.String(p.repository.ParentOwner)
		name = githubv4.String(p.repository.ParentName)
	}

	variables := map[string]interface{}{
		"owner":       owner,
		"name":        name,
		"baseRefName": githubv4.String(p.gh.Spec.Branch),
		"headRefName": githubv4.String(p.gh.HeadBranch),
	}

	err := p.gh.client.Query(context.Background(), &query, variables)
	if err != nil {
		logrus.Debugf("Error getting existing pull-request: %s", err.Error())
		return err
	}

	if len(query.Repository.PullRequests.Nodes) > 0 {
		p.remotePullRequest = query.Repository.PullRequests.Nodes[0]
		// If a remote pullrequest already exist, then we reuse its body to generate the final one
		p.Report = reports.MergeFromString(p.Report, p.remotePullRequest.Body)
		logrus.Debugf("Existing pull-request found: %s", p.remotePullRequest.ID)
	} else {
		logrus.Debugf("No existing pull-request found in repo: %s/%s", owner, name)
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

	owner := githubv4.String(p.repository.Owner)
	repo := githubv4.String(p.repository.Name)

	if p.spec.Parent {
		owner = githubv4.String(p.repository.ParentOwner)
		repo = githubv4.String(p.repository.ParentName)
	}

	variables := map[string]interface{}{
		"owner":      owner,
		"repository": repo,
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
		err := p.gh.client.Query(context.Background(), &query, variables)

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
