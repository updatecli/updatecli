package pullrequest

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	azdoclient "github.com/updatecli/updatecli/pkg/plugins/resources/azuredevops/client"

	azdogit "github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
)

type gitClient interface {
	GetPullRequests(context.Context, azdogit.GetPullRequestsArgs) (*[]azdogit.GitPullRequest, error)
	GetPullRequestIterationChanges(context.Context, azdogit.GetPullRequestIterationChangesArgs) (*azdogit.GitPullRequestIterationChanges, error)
	GetPullRequestIterations(context.Context, azdogit.GetPullRequestIterationsArgs) (*[]azdogit.GitPullRequestIteration, error)
	GetRefs(context.Context, azdogit.GetRefsArgs) (*azdogit.GetRefsResponseValue, error)
	UpdatePullRequest(context.Context, azdogit.UpdatePullRequestArgs) (*azdogit.GitPullRequest, error)
}

const cleanupHeadMatchRetryDelay = time.Second

var (
	// cleanupHeadMatchSleep is a variable to allow overriding time.Sleep in tests.
	cleanupHeadMatchSleep = time.Sleep
)

func (a *AzureDevOps) findExistingPullRequest(ctx context.Context) (*azdogit.GitPullRequest, error) {
	repository, err := a.client.GetRepository(ctx, a.Project, a.Repository)
	if err != nil {
		return nil, fmt.Errorf("find existing pull request: %w", err)
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
		return false, fmt.Errorf("is remote branch exist: %w", err)
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

func (a *AzureDevOps) closePullRequest(ctx context.Context, gitClient gitClient, repositoryID string, pullRequestID int) error {
	status := azdogit.PullRequestStatusValues.Abandoned

	_, err := gitClient.UpdatePullRequest(ctx, azdogit.UpdatePullRequestArgs{
		GitPullRequestToUpdate: &azdogit.GitPullRequest{
			Status: &status,
		},
		Project:       &a.Project,
		RepositoryId:  &repositoryID,
		PullRequestId: &pullRequestID,
	})
	if err != nil {
		return fmt.Errorf("update Azure DevOps pull request: %w", err)
	}

	return nil
}

func (a *AzureDevOps) isPullRequestEmpty(ctx context.Context, gitClient gitClient, repositoryID string, pullRequestID int) (bool, error) {
	latestIteration, err := a.getLatestPullRequestIteration(ctx, gitClient, repositoryID, pullRequestID)
	if err != nil {
		return false, err
	}

	if latestIteration == nil || latestIteration.Id == nil {
		return false, nil
	}

	top := 1
	compareTo := 0

	changes, err := gitClient.GetPullRequestIterationChanges(ctx, azdogit.GetPullRequestIterationChangesArgs{
		Project:       &a.Project,
		RepositoryId:  &repositoryID,
		PullRequestId: &pullRequestID,
		IterationId:   latestIteration.Id,
		Top:           &top,
		CompareTo:     &compareTo,
	})
	if err != nil {
		return false, fmt.Errorf("list Azure DevOps pull request changes: %w", err)
	}

	return changes.ChangeEntries == nil || len(*changes.ChangeEntries) == 0, nil
}

func (a *AzureDevOps) doesPullRequestHeadMatchRemoteBranchHead(ctx context.Context, gitClient gitClient, repositoryID string, pullRequestID int) (bool, error) {
	latestIteration, err := a.getLatestPullRequestIteration(ctx, gitClient, repositoryID, pullRequestID)
	if err != nil {
		return false, err
	}

	if latestIteration == nil || latestIteration.SourceRefCommit == nil || latestIteration.SourceRefCommit.CommitId == nil {
		return false, nil
	}

	sourceBranchRef, err := getBranchRef(ctx, gitClient, a.Project, repositoryID, a.SourceBranch)
	if err != nil {
		return false, err
	}

	if sourceBranchRef == nil || sourceBranchRef.ObjectId == nil {
		return false, nil
	}

	return *latestIteration.SourceRefCommit.CommitId == *sourceBranchRef.ObjectId, nil
}

func (a *AzureDevOps) retryUntilPullRequestHeadMatchesRemoteBranchHead(ctx context.Context, gitClient gitClient, repositoryID string, pullRequestID int, maxAttempts int) (bool, error) {
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		matches, err := a.doesPullRequestHeadMatchRemoteBranchHead(ctx, gitClient, repositoryID, pullRequestID)
		if err != nil {
			return false, err
		}

		if matches {
			return true, nil
		}

		if attempt < maxAttempts {
			logrus.Debugf("Azure DevOps pull request head does not match remote branch head yet, retrying (%d/%d)", attempt, maxAttempts)
			cleanupHeadMatchSleep(cleanupHeadMatchRetryDelay)
		}
	}

	return false, nil
}

func (a *AzureDevOps) getLatestPullRequestIteration(ctx context.Context, gitClient gitClient, repositoryID string, pullRequestID int) (*azdogit.GitPullRequestIteration, error) {
	iterations, err := gitClient.GetPullRequestIterations(ctx, azdogit.GetPullRequestIterationsArgs{
		Project:       &a.Project,
		RepositoryId:  &repositoryID,
		PullRequestId: &pullRequestID,
	})
	if err != nil {
		return nil, fmt.Errorf("list Azure DevOps pull request iterations: %w", err)
	}

	var latestIteration *azdogit.GitPullRequestIteration

	for i := range *iterations {
		iteration := (*iterations)[i]
		if iteration.Id == nil {
			continue
		}

		if latestIteration == nil || *iteration.Id > *latestIteration.Id {
			latestIteration = &iteration
		}
	}

	return latestIteration, nil
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

	return azdoclient.PullRequestURL(a.client.Spec.URL, a.client.Spec.Organization, a.Project, a.Repository, *pr.PullRequestId)
}

func refName(branch string) string {
	if strings.HasPrefix(branch, "refs/") {
		return branch
	}

	return fmt.Sprintf("refs/heads/%s", branch)
}

func branchExists(ctx context.Context, gitClient gitClient, project, repositoryID, branch string) (bool, error) {
	ref, err := getBranchRef(ctx, gitClient, project, repositoryID, branch)
	if err != nil {
		return false, err
	}

	return ref != nil, nil
}

func getBranchRef(ctx context.Context, gitClient gitClient, project, repositoryID, branch string) (*azdogit.GitRef, error) {
	filter := refName(branch)
	top := 1000

	refs, err := gitClient.GetRefs(ctx, azdogit.GetRefsArgs{
		Project:      &project,
		RepositoryId: &repositoryID,
		// For some reason the filter doesn't seem to work
		// We need to retrieve a batch of refs and filter them manually
		//Filter:       &filter,
		Top: &top,
	})
	if err != nil {
		return nil, fmt.Errorf("list Azure DevOps refs for branch %q: %w", branch, err)
	}

	for _, ref := range refs.Value {
		if stringValue(ref.Name) == filter {
			return &ref, nil
		}
	}

	return nil, nil
}

func repositoryID(repository *azdogit.GitRepository) (string, error) {
	if repository == nil || repository.Id == nil {
		return "", fmt.Errorf("azure devops repository not found")
	}

	return repository.Id.String(), nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
