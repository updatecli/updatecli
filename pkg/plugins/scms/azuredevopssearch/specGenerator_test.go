package azuredevopssearch

import (
	"context"
	"testing"

	"github.com/google/uuid"
	azdocore "github.com/microsoft/azure-devops-go-api/azuredevops/v7/core"
	azdogit "github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockClient struct {
	getProjectsFunc     func(context.Context, azdocore.GetProjectsArgs) (*azdocore.GetProjectsResponseValue, error)
	getBranchesFunc     func(context.Context, azdogit.GetBranchesArgs) (*[]azdogit.GitBranchStats, error)
	getRepositoriesFunc func(context.Context, azdogit.GetRepositoriesArgs) (*[]azdogit.GitRepository, error)
}

func (m mockClient) GetProjects(ctx context.Context, args azdocore.GetProjectsArgs) (*azdocore.GetProjectsResponseValue, error) {
	if m.getProjectsFunc == nil {
		return &azdocore.GetProjectsResponseValue{}, nil
	}

	return m.getProjectsFunc(ctx, args)
}

func (m mockClient) GetBranches(ctx context.Context, args azdogit.GetBranchesArgs) (*[]azdogit.GitBranchStats, error) {
	if m.getBranchesFunc == nil {
		return &[]azdogit.GitBranchStats{}, nil
	}

	return m.getBranchesFunc(ctx, args)
}

func (m mockClient) GetRepositories(ctx context.Context, args azdogit.GetRepositoriesArgs) (*[]azdogit.GitRepository, error) {
	if m.getRepositoriesFunc == nil {
		return &[]azdogit.GitRepository{}, nil
	}

	return m.getRepositoriesFunc(ctx, args)
}

func TestScmsGenerator(t *testing.T) {
	t.Run("generates Azure DevOps SCM specs for matching repositories and branches", func(t *testing.T) {
		firstRepositoryID := uuid.New()
		secondRepositoryID := uuid.New()

		search := AzureDevOpsSearch{
			spec: Spec{
				Organization: "updatecli",
				URL:          "https://dev.azure.com",
				Project:      "platform-.*",
				Repository:   "service-.*",
			},
			limit:             DefaultRepositoryLimit,
			branch:            "^(main|release/.+)$",
			projectPattern:    "platform-.*",
			repositoryPattern: "service-.*",
			client: mockClient{
				getProjectsFunc: func(ctx context.Context, args azdocore.GetProjectsArgs) (*azdocore.GetProjectsResponseValue, error) {
					return &azdocore.GetProjectsResponseValue{
						Value: []azdocore.TeamProjectReference{
							{Name: stringPtr("platform-core")},
							{Name: stringPtr("website")},
							{Name: stringPtr("platform-edge")},
						},
					}, nil
				},
				getRepositoriesFunc: func(ctx context.Context, args azdogit.GetRepositoriesArgs) (*[]azdogit.GitRepository, error) {
					switch *args.Project {
					case "platform-core":
						return &[]azdogit.GitRepository{
							{Name: stringPtr("service-api"), Id: &firstRepositoryID},
							{Name: stringPtr("website")},
						}, nil
					case "platform-edge":
						return &[]azdogit.GitRepository{
							{Name: stringPtr("service-worker"), Id: &secondRepositoryID},
						}, nil
					default:
						return &[]azdogit.GitRepository{}, nil
					}
				},
				getBranchesFunc: func(ctx context.Context, args azdogit.GetBranchesArgs) (*[]azdogit.GitBranchStats, error) {
					switch *args.RepositoryId {
					case firstRepositoryID.String():
						require.Equal(t, "platform-core", *args.Project)
						return &[]azdogit.GitBranchStats{
							{Name: stringPtr("refs/heads/main")},
							{Name: stringPtr("refs/heads/release/1.0.0")},
							{Name: stringPtr("refs/heads/develop")},
						}, nil
					case secondRepositoryID.String():
						require.Equal(t, "platform-edge", *args.Project)
						return &[]azdogit.GitBranchStats{
							{Name: stringPtr("refs/heads/main")},
						}, nil
					default:
						return &[]azdogit.GitBranchStats{}, nil
					}
				},
			},
		}

		specs, err := search.ScmsGenerator(context.Background())
		require.NoError(t, err)
		require.Len(t, specs, 3)

		assert.Equal(t, "service-api", specs[0].Repository)
		assert.Equal(t, "platform-core", specs[0].Project)
		assert.Equal(t, "updatecli", specs[0].Organization)
		assert.Equal(t, "main", specs[0].Branch)
		assert.Equal(t, "release/1.0.0", specs[1].Branch)
		assert.Equal(t, "platform-edge", specs[2].Project)
		assert.Equal(t, "service-worker", specs[2].Repository)
	})

	t.Run("honors the repository limit", func(t *testing.T) {
		firstID := uuid.New()
		search := AzureDevOpsSearch{
			spec: Spec{
				Organization: "updatecli",
				URL:          "https://dev.azure.com",
				Project:      "platform-.*",
			},
			limit:             1,
			branch:            "^main$",
			projectPattern:    "platform-.*",
			repositoryPattern: ".*",
			client: mockClient{
				getProjectsFunc: func(ctx context.Context, args azdocore.GetProjectsArgs) (*azdocore.GetProjectsResponseValue, error) {
					return &azdocore.GetProjectsResponseValue{
						Value: []azdocore.TeamProjectReference{
							{Name: stringPtr("platform-core")},
						},
					}, nil
				},
				getRepositoriesFunc: func(ctx context.Context, args azdogit.GetRepositoriesArgs) (*[]azdogit.GitRepository, error) {
					return &[]azdogit.GitRepository{
						{Name: stringPtr("service-a"), Id: &firstID},
					}, nil
				},
				getBranchesFunc: func(ctx context.Context, args azdogit.GetBranchesArgs) (*[]azdogit.GitBranchStats, error) {
					return &[]azdogit.GitBranchStats{
						{Name: stringPtr("refs/heads/main")},
					}, nil
				},
			},
		}

		specs, err := search.ScmsGenerator(context.Background())
		require.NoError(t, err)
		require.Len(t, specs, 1)
		assert.Equal(t, "service-a", specs[0].Repository)
	})

	t.Run("fails on invalid repository regex", func(t *testing.T) {
		search := AzureDevOpsSearch{
			branch:            "^main$",
			projectPattern:    "^platform$",
			repositoryPattern: "[",
		}

		_, err := search.ScmsGenerator(context.Background())
		require.ErrorContains(t, err, "invalid repository regex")
	})
}

func stringPtr(value string) *string {
	return &value
}
