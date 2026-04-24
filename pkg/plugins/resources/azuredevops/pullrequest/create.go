package pullrequest

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/reports"
	utils "github.com/updatecli/updatecli/pkg/plugins/utils/action"

	azdogit "github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
)

// CreateAction opens or updates a pull request on Azure DevOps.
func (a *AzureDevOps) CreateAction(ctx context.Context, report *reports.Action, resetDescription bool) error {
	title := report.Title
	if a.spec.Title != "" {
		title = a.spec.Title
	}

	body, err := utils.GeneratePullRequestBody("", report.ToActionsString())
	if err != nil {
		return fmt.Errorf("generate Azure DevOps pull request body: %w", err)
	}

	if a.spec.Body != "" {
		body = a.spec.Body
	}

	existingPR, err := a.findExistingPullRequest(ctx)
	if err != nil {
		return err
	}

	if existingPR != nil {
		logrus.Debugln("Azure DevOps pull request already exists, updating it")

		mergedDescription := reports.MergeFromString(stringValue(existingPR.Description), report.ToActionsString())
		if a.spec.Body == "" {
			body, err = utils.GeneratePullRequestBody("", mergedDescription)
			if err != nil {
				return fmt.Errorf("generate Azure DevOps pull request body: %w", err)
			}
		}

		updatePR := azdogit.GitPullRequest{
			Title:       &title,
			Description: &body,
		}

		if a.spec.Draft != nil {
			updatePR.IsDraft = a.spec.Draft
		}

		repository, err := a.client.GetRepository(ctx, a.Project, a.Repository)
		if err != nil {
			return fmt.Errorf("get Azure DevOps repository: %w", err)
		}

		repositoryID, err := repositoryID(repository)
		if err != nil {
			return err
		}

		gitClient, err := a.client.NewGitClient(ctx)
		if err != nil {
			return fmt.Errorf("create Azure DevOps git client: %w", err)
		}

		pr, err := gitClient.UpdatePullRequest(ctx, azdogit.UpdatePullRequestArgs{
			GitPullRequestToUpdate: &updatePR,
			Project:                &a.Project,
			RepositoryId:           &repositoryID,
			PullRequestId:          existingPR.PullRequestId,
		})
		if err != nil {
			return fmt.Errorf("update Azure DevOps pull request: %w", err)
		}

		report.Title = stringValue(pr.Title)
		report.Description = stringValue(pr.Description)
		report.Link = a.pullRequestLink(pr)

		return nil
	}

	ok, err := a.isRemoteBranchesExist(ctx)
	if err != nil {
		return err
	}

	if !ok {
		return fmt.Errorf("remote branches %q and %q do not exist, we can't open a pull request", a.SourceBranch, a.TargetBranch)
	}

	repository, err := a.client.GetRepository(ctx, a.Project, a.Repository)
	if err != nil {
		return fmt.Errorf("get Azure DevOps repository: %w", err)
	}

	repositoryID, err := repositoryID(repository)
	if err != nil {
		return err
	}

	sourceRefName := refName(a.SourceBranch)
	targetRefName := refName(a.TargetBranch)

	createPR := azdogit.GitPullRequest{
		Title:         &title,
		Description:   &body,
		SourceRefName: &sourceRefName,
		TargetRefName: &targetRefName,
	}

	if a.spec.Draft != nil {
		createPR.IsDraft = a.spec.Draft
	}

	gitClient, err := a.client.NewGitClient(ctx)
	if err != nil {
		return fmt.Errorf("create Azure DevOps git client: %w", err)
	}

	pr, err := gitClient.CreatePullRequest(ctx, azdogit.CreatePullRequestArgs{
		GitPullRequestToCreate: &createPR,
		Project:                &a.Project,
		RepositoryId:           &repositoryID,
	})
	if err != nil {
		return fmt.Errorf("create Azure DevOps pull request: %w", err)
	}

	report.Title = stringValue(pr.Title)
	report.Description = stringValue(pr.Description)
	report.Link = a.pullRequestLink(pr)

	logrus.Infof("Azure DevOps pull request successfully opened on %q", report.Link)

	return nil
}
