package client

import (
	"context"
	"time"

	azdosdk "github.com/microsoft/azure-devops-go-api/azuredevops/v7"
	azdogit "github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
)

var (
	// DefaultAzureDevOpsURL is the default URL for Azure DevOps organizations.
	DefaultAzureDevOpsURL string = "https://dev.azure.com"
)

type Client struct {
	Spec       Spec
	connection *azdosdk.Connection
}

func New(s Spec) (Client, error) {
	if err := s.Sanitize(); err != nil {
		return Client{}, err
	}

	timeout := 30 * time.Second
	connection := azdosdk.NewPatConnection(s.URL, s.Token)
	connection.Timeout = &timeout

	return Client{
		Spec:       s,
		connection: connection,
	}, nil
}

func (c Client) NewGitClient(ctx context.Context) (azdogit.Client, error) {
	return azdogit.NewClient(ctx, c.connection)
}

func (c Client) GetRepository(ctx context.Context, project, repository string) (*azdogit.GitRepository, error) {
	gitClient, err := c.NewGitClient(ctx)
	if err != nil {
		return nil, err
	}

	return gitClient.GetRepository(ctx, azdogit.GetRepositoryArgs{
		Project:      &project,
		RepositoryId: &repository,
	})
}
