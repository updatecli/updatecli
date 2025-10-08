package github

import (
	"context"
	"fmt"

	"github.com/shurcooL/githubv4"
)

// MockGitHubClient is a mock implementation of the GitHubClient interface
type MockGitHubClient struct {
	mockedQuery interface{}
	mockedErr   error
}

// Query is a mock implementation of the Query method of the GitHubClient interface
func (mock *MockGitHubClient) Query(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	switch q.(type) {
	case *tagsQuery:
		qt, _ := q.(*tagsQuery)
		mt, _ := mock.mockedQuery.(*tagsQuery)
		*qt = *mt
		return mock.mockedErr
	case *releasesQuery:
		qt, _ := q.(*releasesQuery)
		mt, _ := mock.mockedQuery.(*releasesQuery)
		*qt = *mt
		return mock.mockedErr
	case *labelsQuery:
		qt, _ := q.(*labelsQuery)
		mt, _ := mock.mockedQuery.(*labelsQuery)
		*qt = *mt
		return mock.mockedErr
	case *commitQuery:
		qt, _ := q.(*commitQuery)
		mt, _ := mock.mockedQuery.(*commitQuery)
		*qt = *mt
		return mock.mockedErr
	default:
		return fmt.Errorf("mock error: unsupported type for the provided query (%v)", q)
	}
}

// Mutate is a mock implementation of the Mutate method of the GitHubClient interface
func (mock *MockGitHubClient) Mutate(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error {
	return nil
}
