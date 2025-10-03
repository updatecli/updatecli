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
	ErrPullRequestIsInCleanStatus    = errors.New("pull request Pull request is in clean status")
)

// PullRequest contains multiple fields mapped to GitHub V4 api
type PullRequestApi struct {
	ChangedFiles int
	BaseRefName  string
	Body         string
	HeadRefName  string
	ID           string
	State        string
	Title        string
	Url          string
	Number       int32
}

// ActionSpec specifies the configuration of an action of type "GitHub Pull Request"
type ActionSpec struct {
	// automerge allows to enable/disable the automerge feature on new pullrequest
	//
	// compatible:
	//   * action
	//
	// default:
	//   false
	AutoMerge bool `yaml:",omitempty"`
	// title allows to override the pull request title
	//
	// compatible:
	//   * action
	//
	// default:
	//   The default title is fetch from the first following location:
	//   1. The action title
	//   2. The target title if only one target
	//   3. The pipeline target
	//
	Title string `yaml:",omitempty"`
	// description allows to prepend information to the pullrequest description.
	//
	// compatible:
	//   * action
	//
	// default:
	//   empty
	//
	Description string `yaml:",omitempty"`
	// labels specifies repository labels used for the Pull Request.
	//
	// compatible:
	//   * action
	//
	// default:
	//    empty
	//
	// remark:
	//   Labels must already exist on the repository
	//
	Labels []string `yaml:",omitempty"`
	// draft allows to set pull request in draft
	//
	// compatible:
	//   * action
	//
	// default:
	//   false
	Draft bool `yaml:",omitempty"`
	// maintainercannotmodify allows to specify if maintainer can modify pullRequest
	//
	// compatible:
	//   * action
	//
	// default:
	//   false
	MaintainerCannotModify bool `yaml:",omitempty"`
	// mergemethod allows to specifies what merge method is used to incorporate the pull request.
	//
	// compatible:
	//   * action
	//
	// default:
	//   ""
	//
	// remark:
	//   Accept "merge", "squash", "rebase", or ""
	MergeMethod string `yaml:",omitempty"`
	// usetitleforautomerge allows to specifies to use the Pull Request title as commit message when using auto merge,
	//
	// compatible:
	//   * action
	//
	// default:
	//   ""
	//
	// remark:
	//   Only works for "squash" or "rebase"
	UseTitleForAutoMerge bool `yaml:",omitempty"`
	// parent allows to specifies if a pull request should be sent to the parent of the current fork.
	//
	// compatible:
	//   * action
	//
	// default:
	//   false
	//
	Parent bool `yaml:",omitempty"`

	// Reviewers contains the list of assignee to add to the pull request
	// compatible:
	//   * action
	//
	// default: empty
	//
	// remark:
	//   * if reviewer is a team, the format is "organization/team" and the token must have organization read permission.
	Reviewers []string `yaml:",omitempty"`

	// Assignees contains the list of assignee to add to the pull request
	//
	// default: empty
	//
	// remark:
	//   * Please note that contrary to reviewers, assignees only accept GitHub usernames
	Assignees []string `yaml:",omitempty"`
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
	RateLimit RateLimit
}

func NewAction(spec ActionSpec, gh *Github) (PullRequest, error) {
	err := spec.Validate()

	return PullRequest{
		gh:   gh,
		spec: spec,
	}, err
}

// CleanAction verifies if an existing action requires some cleanup such as closing a pullrequest with no changes.
func (p *PullRequest) CleanAction(report *reports.Action) error {

	repository, err := p.gh.queryRepository("", "")
	if err != nil {
		return err
	}

	p.repository = repository

	// Check if there is already a pullRequest for current pipeline
	err = p.getRemotePullRequest(false, 0)
	if err != nil {
		return err
	}

	if p.remotePullRequest.ID == "" {
		logrus.Debugln("nothing to clean")
		return nil
	}

	if p.remotePullRequest.ChangedFiles == 0 {
		logrus.Debugf("No changed file detected at pull request:\n\t%s", p.remotePullRequest.Url)
		// Not returning an error if the comment failed to be added
		// as the main purpose of this function is to close the pullrequest
		err = p.closePullRequest(0)
		if err != nil {
			return fmt.Errorf("closing pull request: %w", err)
		}

		report.Link = ""
		report.Description = "Pull request closed as no changed file detected"

		return nil
	}

	return nil
}

func (p *PullRequest) CheckActionExist(report *reports.Action) error {

	repository, err := p.gh.queryRepository("", "")
	if err != nil {
		return fmt.Errorf("querying repository: %w", err)
	}

	p.repository = repository

	err = p.getRemotePullRequest(false, 0)
	if err != nil {
		return fmt.Errorf("error getting remote pull request: %w", err)
	}

	if p.remotePullRequest.ID == "" {
		return nil
	}

	report.Link = p.remotePullRequest.Url
	report.Title = p.remotePullRequest.Title
	report.Description = p.remotePullRequest.Body

	return nil
}

// CreateAction creates a new GitHub Pull Request or update an existing one.
func (p *PullRequest) CreateAction(report *reports.Action, resetDescription bool) error {

	// One GitHub pullrequest body can contain multiple action report
	// It would be better to refactor CreateAction
	p.Report = report.ToActionsString()
	p.Title = report.Title

	if p.spec.Title != "" {
		p.Title = p.spec.Title
	}

	sourceBranch, workingBranch, _ := p.gh.GetBranches()

	repository, err := p.gh.queryRepository(sourceBranch, workingBranch)
	if err != nil {
		return err
	}

	p.repository = repository

	// Check if there is already a pullRequest for current pipeline
	err = p.getRemotePullRequest(resetDescription, 0)
	if err != nil {
		return err
	}

	// If we didn't find a Pull Request ID then it means we need to create a new pullrequest.
	if len(p.remotePullRequest.ID) == 0 {
		if err := p.OpenPullRequest(0); err != nil {
			return err
		}
	}

	if len(p.remotePullRequest.ID) == 0 {
		return nil
	}

	report.Link = p.remotePullRequest.Url
	report.Title = p.remotePullRequest.Title
	report.Description = p.remotePullRequest.Body

	// Once the remote Pull Request exists, we can than update it with additional information such as
	// tags,assignee,etc.
	if err := p.updatePullRequest(0); err != nil {
		return err
	}

	if p.spec.AutoMerge {
		if err := p.EnablePullRequestAutoMerge(0); err != nil {
			switch err.Error() {
			case ErrAutomergeNotAllowOnRepository.Error():
				logrus.Errorln("Automerge can't be enabled. Make sure to all it on the repository.")
			case ErrPullRequestIsInCleanStatus.Error():
				logrus.Errorln("Automerge can't be enabled. Make sure to have branch protection rules enabled on the repository.")
			default:
				logrus.Debugf("Error enabling automerge: %s", err.Error())
			}
			return err
		}
	}

	return nil
}

// closePullRequest closes an existing Pull Request using GitHub graphql api.
func (p *PullRequest) closePullRequest(retry int) error {

	// https://docs.github.com/en/graphql/reference/input-objects#closepullrequestinput
	/*
		  mutation($input: closePullRequestInput!){
			closePullRequest(input:$input){
			  pullRequest{
				url
			    title
			    body
			  }
			}
		  }

		  {
			"input": {
			  "pullRequestId" : "yyy"
			}
		  }
	*/
	var mutation struct {
		UpdatePullRequest struct {
			PullRequest PullRequestApi
		} `graphql:"closePullRequest(input: $input)"`
		RateLimit RateLimit
	}

	input := githubv4.ClosePullRequestInput{
		PullRequestID: githubv4.ID(p.remotePullRequest.ID),
	}

	err := p.gh.client.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		logrus.Debugf("Closing pull request: %s", err.Error())
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			logrus.Debugln(mutation.RateLimit)
			if retry < MaxRetry {
				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, MaxRetry)
				return p.closePullRequest(retry + 1)
			}
			return fmt.Errorf("%s", ErrAPIRateLimitExceededFinalAttempt)
		}
		return fmt.Errorf("closing pull request: %w", err)
	}

	logrus.Debugln(mutation.RateLimit)

	msg := "Pull request closed as no changed file detected"
	logrus.Infof("%s at:\n\n\t%s\n\n", msg, mutation.UpdatePullRequest.PullRequest.Url)
	err = p.addComment(msg, 0)
	if err != nil {
		logrus.Errorf("Commenting pull-request: %s", err.Error())
	}

	return nil
}

// updatePullRequest updates an existing Pull Request.
func (p *PullRequest) updatePullRequest(retry int) error {

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
		RateLimit RateLimit
	}

	logrus.Debugf("Updating GitHub Pull Request")

	title := p.Title

	bodyPR, err := utils.GeneratePullRequestBody(p.spec.Description, p.Report)
	if err != nil {
		return err
	}

	labelsID := []githubv4.ID{}
	repositoryLabels, err := p.gh.getRepositoryLabels(0)
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

	remotePRLabels, err := p.GetPullRequestLabelsInformation(0)
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

	if len(p.spec.Reviewers) != 0 {
		err = p.addPullrequestReviewers(p.remotePullRequest.ID, 0)
		if err != nil {
			logrus.Debugln(err.Error())
		}
	}

	if len(p.spec.Assignees) != 0 {
		var assigneesID []githubv4.ID
		for _, assignee := range p.spec.Assignees {
			user, err := getUserInfo(p.gh.client, assignee, 0)
			if err != nil {
				logrus.Debugf("Failed to get user id for %s: %v", assignee, err)
				continue
			}

			logrus.Debugf("Assignee ID %q found for %q", user.ID, assignee)
			assigneesID = append(assigneesID, githubv4.NewID(user.ID))
		}

		if len(assigneesID) > 0 {
			input.AssigneeIDs = &assigneesID
		}
	}

	if len(p.spec.Labels) != 0 {
		input.LabelIDs = &labelsID
	}

	err = p.gh.client.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			logrus.Debugln(mutation.RateLimit)
			if retry < MaxRetry {
				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, MaxRetry)
				return p.updatePullRequest(retry + 1)
			}
			return fmt.Errorf("%s", ErrAPIRateLimitExceededFinalAttempt)
		}
		logrus.Debugf("Error updating pull-request: %s", err.Error())
		return fmt.Errorf("updating pull request: %w", err)
	}

	logrus.Debugln(mutation.RateLimit)

	logrus.Infof("\nPull Request available at:\n\n\t%s\n\n", mutation.UpdatePullRequest.PullRequest.Url)

	return nil
}

// EnablePullRequestAutoMerge updates an existing pullrequest with the flag automerge
func (p *PullRequest) EnablePullRequestAutoMerge(retry int) error {

	// Test that automerge feature is enabled on repository but only if we plan to use it
	autoMergeAllowed, err := p.isAutoMergedEnabledOnRepository(0)
	if err != nil {
		return err
	}

	if !autoMergeAllowed {
		return ErrAutomergeNotAllowOnRepository
	}

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

	var mutation mutationEnablePullRequestAutoMerge
	err = p.gh.client.Mutate(context.Background(), &mutation, input, nil)

	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			logrus.Debugln(mutation.RateLimit)
			if retry < MaxRetry {
				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, MaxRetry)
				return p.EnablePullRequestAutoMerge(retry + 1)
			}
			return fmt.Errorf("%s", ErrAPIRateLimitExceededFinalAttempt)
		}
		return fmt.Errorf("enabling pull request automerge: %w", err)
	}

	logrus.Debugln(mutation.RateLimit)

	return nil
}

// OpenPullRequest creates a new GitHub Pull Request.
func (p *PullRequest) OpenPullRequest(retry int) error {

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
	bodyPR, err := utils.GeneratePullRequestBody(p.spec.Description, p.Report)
	if err != nil {
		return err
	}

	sourceBranch, workingBranch, targetBranch := p.gh.GetBranches()

	logrus.Debugf("Branch %s is %s of %s", workingBranch, strings.ToLower(p.repository.Status), sourceBranch)

	// Check if they are changes that need to be published otherwise exit
	// It's worth mentioning that at this time, changes have already been published
	// The goal is just to not open a pull request if there is no changes
	// https://docs.github.com/en/graphql/reference/enums#comparisonstatus
	switch p.repository.Status {
	case "AHEAD":
		logrus.Debugf("Opening GitHub pull request")
	case "DIVERGED":
		logrus.Debugf("Opening GitHub pull request even though the branch is diverged")
	case "BEHIND", "IDENTICAL":
		logrus.Debugf("GitHub pull request not needed")
		return nil
	default:
		return fmt.Errorf("unknown status %q", p.repository.Status)
	}

	input := githubv4.CreatePullRequestInput{
		RepositoryID:        githubv4.String(p.repository.ID),
		BaseRefName:         githubv4.String(targetBranch),
		HeadRefName:         githubv4.String(workingBranch),
		Title:               githubv4.String(p.Title),
		Body:                githubv4.NewString(githubv4.String(bodyPR)),
		MaintainerCanModify: githubv4.NewBoolean(githubv4.Boolean(!p.spec.MaintainerCannotModify)),
		Draft:               githubv4.NewBoolean(githubv4.Boolean(p.spec.Draft)),
	}

	if p.spec.Parent {
		input.RepositoryID = githubv4.String(p.repository.ParentID)
		input.HeadRepositoryID = githubv4.NewID(p.repository.ID)
	}

	logrus.Debugf("Opening pull-request against repo from %q to %q", input.BaseRefName, input.HeadRefName)

	var mutation struct {
		CreatePullRequest struct {
			PullRequest PullRequestApi
		} `graphql:"createPullRequest(input: $input)"`
		RateLimit RateLimit
	}

	err = p.gh.client.Mutate(context.Background(), &mutation, input, nil)
	if err != nil {
		logrus.Infof("\nError creating pull request:\n\n\t%s\n\n", err.Error())
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			logrus.Debugln(mutation.RateLimit)
			if retry < MaxRetry {
				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, MaxRetry)
				return p.OpenPullRequest(retry + 1)
			}
			return fmt.Errorf("%s", ErrAPIRateLimitExceededFinalAttempt)
		}
		return fmt.Errorf("creating pull request: %w", err)
	}

	logrus.Debugln(mutation.RateLimit)

	p.remotePullRequest = mutation.CreatePullRequest.PullRequest

	return nil

}

// isAutoMergedEnabledOnRepository checks if a remote repository allows automerging Pull Requests.
func (p *PullRequest) isAutoMergedEnabledOnRepository(retry int) (bool, error) {

	var query struct {
		Repository struct {
			AutoMergeAllowed bool
		} `graphql:"repository(owner: $owner, name: $name)"`
		RateLimit RateLimit
	}

	variables := map[string]interface{}{
		"owner": githubv4.String(p.gh.Spec.Owner),
		"name":  githubv4.String(p.gh.Spec.Repository),
	}

	err := p.gh.client.Query(context.Background(), &query, variables)

	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			logrus.Debugln(query.RateLimit)
			if retry < MaxRetry {
				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, MaxRetry)
				return p.isAutoMergedEnabledOnRepository(retry + 1)
			}
			return false, fmt.Errorf("%s", ErrAPIRateLimitExceededFinalAttempt)
		}
		return false, fmt.Errorf("querying GitHub API: %w", err)
	}
	logrus.Debugln(query.RateLimit)

	return query.Repository.AutoMergeAllowed, nil

}

// getRemotePullRequest checks if a Pull Request already exists on GitHub and is in the state 'open' or 'closed'.
func (p *PullRequest) getRemotePullRequest(resetBody bool, retry int) error {
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

	_, workingBranch, targetBranch := p.gh.GetBranches()

	variables := map[string]interface{}{
		"owner":       owner,
		"name":        name,
		"baseRefName": githubv4.String(targetBranch),
		"headRefName": githubv4.String(workingBranch),
	}

	err := p.gh.client.Query(context.Background(), &query, variables)
	if err != nil {
		if strings.Contains(err.Error(), ErrAPIRateLimitExceeded) {
			logrus.Debugln(query.Repository)
			if retry < MaxRetry {
				logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, MaxRetry)
				return p.getRemotePullRequest(resetBody, retry+1)
			}
			return fmt.Errorf("%s", ErrAPIRateLimitExceededFinalAttempt)
		}
		return fmt.Errorf("getting existing pull request: %w", err)
	}

	// If no pull-request found, then we can exit
	if len(query.Repository.PullRequests.Nodes) == 0 {
		logrus.Debugf("No existing pull-request found in repo: %s/%s", owner, name)
		return nil
	}

	p.remotePullRequest = query.Repository.PullRequests.Nodes[0]
	// If a remote pullrequest already exist, then we reuse its body to generate the final one
	switch resetBody {
	case false:
		logrus.Debugf("Merging existing pull-request body with new report")
		p.Report = reports.MergeFromString(p.remotePullRequest.Body, p.Report)
	case true:
		logrus.Debugf("Resetting pull-request body with new report")
	}

	logrus.Infof("Existing GitHub pull request found: %s", p.remotePullRequest.Url)

	return nil
}

// getPullRequestLabelsInformation queries GitHub Api to retrieve every labels assigned to a pullRequest
func (p *PullRequest) GetPullRequestLabelsInformation(retry int) ([]repositoryLabelApi, error) {

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
			if strings.Contains(err.Error(), "API rate limit exceeded") {
				logrus.Debugln(query.RateLimit)
				if retry < MaxRetry {
					query.RateLimit.Pause()

					logrus.Warningf("GitHub API rate limit exceeded. Retrying... (%d/%d)", retry+1, MaxRetry)
					return p.GetPullRequestLabelsInformation(retry + 1)
				}
				return nil, fmt.Errorf("%s", ErrAPIRateLimitExceededFinalAttempt)
			}

			return nil, fmt.Errorf("querying GitHub API: %w", err)
		}

		logrus.Debugln(query.RateLimit)

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
