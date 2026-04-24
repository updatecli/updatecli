package pullrequest

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	azdoclient "github.com/updatecli/updatecli/pkg/plugins/resources/azuredevops/client"

	azdogit "github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
)

func (a *AzureDevOps) findExistingPullRequest(ctx context.Context) (*azdogit.GitPullRequest, error) {
	repository, err := a.client.GetRepository(ctx, a.Project, a.Repository)
	if err != nil {
		return nil, fmt.Errorf("get Azure DevOps repository: %w", err)
	}

	repositoryID, err := repositoryID(repository)
	if err != nil {
		return nil, err
	}

	gitClient, err := a.client.NewGitClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create Azure DevOps git client: %w", err)
	}

	sourceRefName := refName(a.SourceBranch)
	targetRefName := refName(a.TargetBranch)
	status := azdogit.PullRequestStatusValues.Active

	pullRequests, err := gitClient.GetPullRequests(ctx, azdogit.GetPullRequestsArgs{
		Project:      &a.Project,
		RepositoryId: &repositoryID,
		SearchCriteria: &azdogit.GitPullRequestSearchCriteria{
			RepositoryId:  repository.Id,
			SourceRefName: &sourceRefName,
			Status:        &status,
			TargetRefName: &targetRefName,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("list Azure DevOps pull requests: %w", err)
	}

	for _, pr := range *pullRequests {
		if stringValue(pr.SourceRefName) == sourceRefName &&
			stringValue(pr.TargetRefName) == targetRefName &&
			pr.Status != nil &&
			*pr.Status == azdogit.PullRequestStatusValues.Active {
			logrus.Infof("%s Azure DevOps pull request detected at:\n\t%s",
				result.SUCCESS,
				a.pullRequestLink(&pr))

			return &pr, nil
		}
	}

	return nil, nil
}

func (a *AzureDevOps) isRemoteBranchesExist(ctx context.Context) (bool, error) {
	repository, err := a.client.GetRepository(ctx, a.Project, a.Repository)
	if err != nil {
		return false, fmt.Errorf("get Azure DevOps repository: %w", err)
	}

	repositoryID, err := repositoryID(repository)
	if err != nil {
		return false, err
	}

	gitClient, err := a.client.NewGitClient(ctx)
	if err != nil {
		return false, fmt.Errorf("create Azure DevOps git client: %w", err)
	}

	sourceExists, err := branchExists(ctx, gitClient, a.Project, repositoryID, a.SourceBranch)
	if err != nil {
		return false, err
	}

	targetExists, err := branchExists(ctx, gitClient, a.Project, repositoryID, a.TargetBranch)
	if err != nil {
		return false, err
	}

	if !sourceExists {
		logrus.Debugf("Branch %q not found on remote repository %s/%s", a.SourceBranch, a.Project, a.Repository)
	}

	if !targetExists {
		logrus.Debugf("Branch %q not found on remote repository %s/%s", a.TargetBranch, a.Project, a.Repository)
	}

	return sourceExists && targetExists, nil
}

func (a *AzureDevOps) inheritFromScm() {
	if a.scm != nil {
		_, a.SourceBranch, a.TargetBranch = a.scm.GetBranches()
		a.Project = a.scm.Spec.Project
		a.Repository = a.scm.Spec.Repository
	}

	if a.spec.SourceBranch != "" {
		a.SourceBranch = a.spec.SourceBranch
	}

	if a.spec.TargetBranch != "" {
		a.TargetBranch = a.spec.TargetBranch
	}

	if a.spec.Project != "" {
		a.Project = a.spec.Project
	}

	if a.spec.Repository != "" {
		a.Repository = a.spec.Repository
	}
}

func (a *AzureDevOps) pullRequestLink(pr *azdogit.GitPullRequest) string {
	if pr == nil || pr.PullRequestId == nil {
		return ""
	}

	return azdoclient.PullRequestURL(a.client.Spec.URL, a.Project, a.Repository, *pr.PullRequestId)
}

func refName(branch string) string {
	if strings.HasPrefix(branch, "refs/") {
		return branch
	}

	return fmt.Sprintf("refs/heads/%s", branch)
}

func branchExists(ctx context.Context, gitClient azdogit.Client, project, repositoryID, branch string) (bool, error) {
	filter := refName(branch)
	top := 1

	refs, err := gitClient.GetRefs(ctx, azdogit.GetRefsArgs{
		Project:      &project,
		RepositoryId: &repositoryID,
		Filter:       &filter,
		Top:          &top,
	})
	if err != nil {
		return false, fmt.Errorf("list Azure DevOps refs for branch %q: %w", branch, err)
	}

	for _, ref := range refs.Value {
		if stringValue(ref.Name) == filter {
			return true, nil
		}
	}

	return false, nil
}

func repositoryID(repository *azdogit.GitRepository) (string, error) {
	if repository == nil || repository.Id == nil {
		return "", fmt.Errorf("azure devops repository ID not found")
	}

	return repository.Id.String(), nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
