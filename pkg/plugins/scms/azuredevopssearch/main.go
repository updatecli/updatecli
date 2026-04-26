package azuredevopssearch

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	azdocore "github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	azdogit "github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	azclient "github.com/updatecli/updatecli/pkg/plugins/resources/azuredevops/client"
)

const Kind = "azuredevopssearch"

type client interface {
	GetProjects(context.Context, azdocore.GetProjectsArgs) (*azdocore.GetProjectsResponseValue, error)
	GetBranches(context.Context, azdogit.GetBranchesArgs) (*[]azdogit.GitBranchStats, error)
	GetRepositories(context.Context, azdogit.GetRepositoriesArgs) (*[]azdogit.GitRepository, error)
}

type AzureDevOpsSearch struct {
	spec              Spec
	limit             int
	branch            string
	projectPattern    string
	repositoryPattern string
	client            client
}

func New(s interface{}) (*AzureDevOpsSearch, error) {
	var spec Spec

	if err := mapstructure.Decode(s, &spec); err != nil {
		return nil, err
	}

	spec.sanitize()
	if err := spec.Validate(); err != nil {
		return nil, err
	}

	limit := DefaultRepositoryLimit
	if spec.Limit != nil {
		limit = *spec.Limit
	}

	branch := "^main$"
	if spec.Branch != "" {
		branch = spec.Branch
	}

	if _, err := regexp.Compile(spec.Project); err != nil {
		return nil, fmt.Errorf("invalid project regex %q: %w", spec.Project, err)
	}

	repositoryPattern := ".*"
	if spec.Repository != "" {
		repositoryPattern = spec.Repository
	}

	if _, err := regexp.Compile(repositoryPattern); err != nil {
		return nil, fmt.Errorf("invalid repository regex %q: %w", repositoryPattern, err)
	}

	if _, err := regexp.Compile(branch); err != nil {
		return nil, fmt.Errorf("invalid branch regex %q: %w", branch, err)
	}

	client := azureDevOpsClient{}

	azureDevOpsClient, err := azclient.New(azclient.Spec{
		URL:          spec.URL,
		Organization: spec.Organization,
		Project:      spec.Project,
		Repository:   spec.Repository,
		Token:        spec.Token,
		Username:     spec.Username,
	})
	if err != nil {
		return nil, fmt.Errorf("creating Azure DevOps client: %w", err)
	}

	coreClient, err := azureDevOpsClient.NewCoreClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("creating Azure DevOps core client: %w", err)
	}

	gitClient, err := azureDevOpsClient.NewGitClient(context.Background())
	if err != nil {
		return nil, fmt.Errorf("creating Azure DevOps git client: %w", err)
	}

	client.core = coreClient
	client.git = gitClient

	return &AzureDevOpsSearch{
		spec:              spec,
		limit:             limit,
		branch:            branch,
		projectPattern:    spec.Project,
		repositoryPattern: repositoryPattern,
		client:            client,
	}, nil
}

type azureDevOpsClient struct {
	core azdocore.Client
	git  azdogit.Client
}

func (c azureDevOpsClient) GetProjects(ctx context.Context, args azdocore.GetProjectsArgs) (*azdocore.GetProjectsResponseValue, error) {
	return c.core.GetProjects(ctx, args)
}

func (c azureDevOpsClient) GetBranches(ctx context.Context, args azdogit.GetBranchesArgs) (*[]azdogit.GitBranchStats, error) {
	return c.git.GetBranches(ctx, args)
}

func (c azureDevOpsClient) GetRepositories(ctx context.Context, args azdogit.GetRepositoriesArgs) (*[]azdogit.GitRepository, error) {
	return c.git.GetRepositories(ctx, args)
}

func normalizeBranchName(branch string) string {
	return strings.TrimPrefix(branch, "refs/heads/")
}
