package azuredevopssearch

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	azdocore "github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	azdogit "github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/sirupsen/logrus"
	azdoclient "github.com/updatecli/updatecli/pkg/plugins/resources/azuredevops/client"
	azdoscm "github.com/updatecli/updatecli/pkg/plugins/scms/azuredevops"
)

// ScmsGenerator discovers Azure DevOps repositories within the configured project
// and returns a list of azuredevops.Spec, one per matching branch.
func (a *AzureDevOpsSearch) ScmsGenerator(ctx context.Context) ([]azdoscm.Spec, error) {
	results := make([]azdoscm.Spec, 0)

	branchRegex, err := regexp.Compile(a.branch)
	if err != nil {
		return nil, fmt.Errorf("invalid branch regex %q: %w", a.branch, err)
	}

	projectRegex, err := regexp.Compile(a.projectPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid project regex %q: %w", a.projectPattern, err)
	}

	repositoryRegex, err := regexp.Compile(a.repositoryPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid repository regex %q: %w", a.repositoryPattern, err)
	}

	projects, err := a.listProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list Azure DevOps projects for organization %q: %w", a.spec.Organization, err)
	}

	for _, project := range projects {
		projectName := stringValue(project.Name)
		if projectName == "" || !projectRegex.MatchString(projectName) {
			continue
		}

		repositories, err := a.client.GetRepositories(ctx, azdogit.GetRepositoriesArgs{
			Project: &projectName,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list Azure DevOps repositories for project %q: %w", projectName, err)
		}

		for _, repository := range *repositories {
			if repository.IsDisabled != nil && *repository.IsDisabled {
				continue
			}

			repositoryName := stringValue(repository.Name)
			if repositoryName == "" {
				logrus.Warning("Skipping Azure DevOps repository with no name")
				continue
			}

			if !repositoryRegex.MatchString(repositoryName) {
				continue
			}

			if repository.Id == nil {
				logrus.Warningf("Skipping Azure DevOps repository %q with no ID", repositoryName)
				continue
			}

			logrus.Debugf("Processing Azure DevOps repository: %s/%s", projectName, repositoryName)

			branches, err := a.listBranches(ctx, projectName, repository.Id.String())
			if err != nil {
				return nil, fmt.Errorf("failed to list branches for Azure DevOps repository %q in project %q: %w", repositoryName, projectName, err)
			}

			for _, branchName := range branches {
				if !branchRegex.MatchString(branchName) {
					continue
				}

				spec := azdoscm.Spec{
					Branch:                 branchName,
					CommitMessage:          a.spec.CommitMessage,
					Directory:              a.spec.Directory,
					Depth:                  a.spec.Depth,
					Email:                  a.spec.Email,
					Force:                  a.spec.Force,
					GPG:                    a.spec.GPG,
					Submodules:             a.spec.Submodules,
					User:                   a.spec.User,
					WorkingBranch:          a.spec.WorkingBranch,
					WorkingBranchPrefix:    a.spec.WorkingBranchPrefix,
					WorkingBranchSeparator: a.spec.WorkingBranchSeparator,
					Spec: azdoclient.Spec{
						Organization: a.spec.Organization,
						URL:          a.spec.URL,
						Project:      projectName,
						Repository:   repositoryName,
						Token:        a.spec.Token,
						Username:     a.spec.Username,
					},
				}

				results = append(results, spec)

				if a.limit > 0 && len(results) >= a.limit {
					return results, nil
				}
			}
		}
	}

	return results, nil
}

func (a *AzureDevOpsSearch) listProjects(ctx context.Context) ([]azdocore.TeamProjectReference, error) {
	results := make([]azdocore.TeamProjectReference, 0)
	top := 100
	var continuationToken *int

	for {
		response, err := a.client.GetProjects(ctx, azdocore.GetProjectsArgs{
			Top:               &top,
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return nil, err
		}

		results = append(results, response.Value...)

		if response.ContinuationToken == "" {
			break
		}

		nextToken, err := strconv.Atoi(response.ContinuationToken)
		if err != nil {
			return nil, fmt.Errorf("parse Azure DevOps continuation token %q: %w", response.ContinuationToken, err)
		}
		continuationToken = &nextToken
	}

	return results, nil
}

func (a *AzureDevOpsSearch) listBranches(ctx context.Context, projectName, repositoryID string) ([]string, error) {
	branches, err := a.client.GetBranches(ctx, azdogit.GetBranchesArgs{
		Project:      &projectName,
		RepositoryId: &repositoryID,
	})
	if err != nil {
		return nil, err
	}

	results := make([]string, 0, len(*branches))
	for _, branch := range *branches {
		branchName := normalizeBranchName(stringValue(branch.Name))
		if branchName == "" {
			continue
		}

		results = append(results, branchName)
	}

	return results, nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return *value
}
