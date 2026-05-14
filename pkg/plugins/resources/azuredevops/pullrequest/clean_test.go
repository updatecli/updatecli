package pullrequest

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	azdogit "github.com/microsoft/azure-devops-go-api/azuredevops/v7/git"
)

type mockGitClient struct {
	getPullRequestsFunc                func(context.Context, azdogit.GetPullRequestsArgs) (*[]azdogit.GitPullRequest, error)
	getPullRequestIterationChangesFunc func(context.Context, azdogit.GetPullRequestIterationChangesArgs) (*azdogit.GitPullRequestIterationChanges, error)
	getPullRequestIterationsFunc       func(context.Context, azdogit.GetPullRequestIterationsArgs) (*[]azdogit.GitPullRequestIteration, error)
	getRefsFunc                        func(context.Context, azdogit.GetRefsArgs) (*azdogit.GetRefsResponseValue, error)
	updatePullRequestFunc              func(context.Context, azdogit.UpdatePullRequestArgs) (*azdogit.GitPullRequest, error)
}

func (m mockGitClient) GetPullRequests(ctx context.Context, args azdogit.GetPullRequestsArgs) (*[]azdogit.GitPullRequest, error) {
	if m.getPullRequestsFunc == nil {
		return &[]azdogit.GitPullRequest{}, nil
	}

	return m.getPullRequestsFunc(ctx, args)
}

func (m mockGitClient) GetPullRequestIterationChanges(ctx context.Context, args azdogit.GetPullRequestIterationChangesArgs) (*azdogit.GitPullRequestIterationChanges, error) {
	if m.getPullRequestIterationChangesFunc == nil {
		return &azdogit.GitPullRequestIterationChanges{}, nil
	}

	return m.getPullRequestIterationChangesFunc(ctx, args)
}

func (m mockGitClient) GetPullRequestIterations(ctx context.Context, args azdogit.GetPullRequestIterationsArgs) (*[]azdogit.GitPullRequestIteration, error) {
	if m.getPullRequestIterationsFunc == nil {
		return &[]azdogit.GitPullRequestIteration{}, nil
	}

	return m.getPullRequestIterationsFunc(ctx, args)
}

func (m mockGitClient) GetRefs(ctx context.Context, args azdogit.GetRefsArgs) (*azdogit.GetRefsResponseValue, error) {
	if m.getRefsFunc == nil {
		return &azdogit.GetRefsResponseValue{}, nil
	}

	return m.getRefsFunc(ctx, args)
}

func (m mockGitClient) UpdatePullRequest(ctx context.Context, args azdogit.UpdatePullRequestArgs) (*azdogit.GitPullRequest, error) {
	if m.updatePullRequestFunc == nil {
		return &azdogit.GitPullRequest{}, nil
	}

	return m.updatePullRequestFunc(ctx, args)
}

func TestIsPullRequestEmpty(t *testing.T) {
	t.Run("returns true when latest iteration has no changes", func(t *testing.T) {
		pr := AzureDevOps{Project: "project"}
		pullRequestID := 42

		isEmpty, err := pr.isPullRequestEmpty(context.Background(), mockGitClient{
			getPullRequestIterationsFunc: func(ctx context.Context, args azdogit.GetPullRequestIterationsArgs) (*[]azdogit.GitPullRequestIteration, error) {
				require.Equal(t, pullRequestID, *args.PullRequestId)
				return &[]azdogit.GitPullRequestIteration{
					{Id: intPtr(1)},
					{Id: intPtr(3)},
					{Id: intPtr(2)},
				}, nil
			},
			getPullRequestIterationChangesFunc: func(ctx context.Context, args azdogit.GetPullRequestIterationChangesArgs) (*azdogit.GitPullRequestIterationChanges, error) {
				require.Equal(t, 3, *args.IterationId)
				return &azdogit.GitPullRequestIterationChanges{
					ChangeEntries: &[]azdogit.GitPullRequestChange{},
				}, nil
			},
		}, "repository-id", pullRequestID)

		require.NoError(t, err)
		assert.True(t, isEmpty)
	})

	t.Run("returns false when latest iteration has changes", func(t *testing.T) {
		pr := AzureDevOps{Project: "project"}
		pullRequestID := 42

		isEmpty, err := pr.isPullRequestEmpty(context.Background(), mockGitClient{
			getPullRequestIterationsFunc: func(ctx context.Context, args azdogit.GetPullRequestIterationsArgs) (*[]azdogit.GitPullRequestIteration, error) {
				return &[]azdogit.GitPullRequestIteration{
					{Id: intPtr(1)},
				}, nil
			},
			getPullRequestIterationChangesFunc: func(ctx context.Context, args azdogit.GetPullRequestIterationChangesArgs) (*azdogit.GitPullRequestIterationChanges, error) {
				return &azdogit.GitPullRequestIterationChanges{
					ChangeEntries: &[]azdogit.GitPullRequestChange{{}},
				}, nil
			},
		}, "repository-id", pullRequestID)

		require.NoError(t, err)
		assert.False(t, isEmpty)
	})
}

func TestClosePullRequest(t *testing.T) {
	pr := AzureDevOps{Project: "project"}
	pullRequestID := 7

	err := pr.closePullRequest(context.Background(), mockGitClient{
		updatePullRequestFunc: func(ctx context.Context, args azdogit.UpdatePullRequestArgs) (*azdogit.GitPullRequest, error) {
			require.Equal(t, "repository-id", *args.RepositoryId)
			require.Equal(t, "project", *args.Project)
			require.Equal(t, pullRequestID, *args.PullRequestId)
			require.NotNil(t, args.GitPullRequestToUpdate)
			require.NotNil(t, args.GitPullRequestToUpdate.Status)
			assert.Equal(t, azdogit.PullRequestStatusValues.Abandoned, *args.GitPullRequestToUpdate.Status)

			return &azdogit.GitPullRequest{}, nil
		},
	}, "repository-id", pullRequestID)

	require.NoError(t, err)
}

func TestDoesPullRequestHeadMatchRemoteBranchHead(t *testing.T) {
	t.Run("returns true when latest iteration head matches remote branch head", func(t *testing.T) {
		pr := AzureDevOps{
			Project:      "project",
			SourceBranch: "main",
		}
		pullRequestID := 42
		commitID := "abc123"

		matches, err := pr.doesPullRequestHeadMatchRemoteBranchHead(context.Background(), mockGitClient{
			getPullRequestIterationsFunc: func(ctx context.Context, args azdogit.GetPullRequestIterationsArgs) (*[]azdogit.GitPullRequestIteration, error) {
				return &[]azdogit.GitPullRequestIteration{
					{Id: intPtr(1)},
					{
						Id: intPtr(2),
						SourceRefCommit: &azdogit.GitCommitRef{
							CommitId: stringPtr(commitID),
						},
					},
				}, nil
			},
			getRefsFunc: func(ctx context.Context, args azdogit.GetRefsArgs) (*azdogit.GetRefsResponseValue, error) {
				require.Equal(t, "repository-id", *args.RepositoryId)
				require.Equal(t, "project", *args.Project)
				return &azdogit.GetRefsResponseValue{
					Value: []azdogit.GitRef{
						{
							Name:     stringPtr("refs/heads/main"),
							ObjectId: stringPtr(commitID),
						},
					},
				}, nil
			},
		}, "repository-id", pullRequestID)

		require.NoError(t, err)
		assert.True(t, matches)
	})

	t.Run("returns false when latest iteration head does not match remote branch head", func(t *testing.T) {
		pr := AzureDevOps{
			Project:      "project",
			SourceBranch: "main",
		}

		matches, err := pr.doesPullRequestHeadMatchRemoteBranchHead(context.Background(), mockGitClient{
			getPullRequestIterationsFunc: func(ctx context.Context, args azdogit.GetPullRequestIterationsArgs) (*[]azdogit.GitPullRequestIteration, error) {
				return &[]azdogit.GitPullRequestIteration{
					{
						Id: intPtr(2),
						SourceRefCommit: &azdogit.GitCommitRef{
							CommitId: stringPtr("abc123"),
						},
					},
				}, nil
			},
			getRefsFunc: func(ctx context.Context, args azdogit.GetRefsArgs) (*azdogit.GetRefsResponseValue, error) {
				return &azdogit.GetRefsResponseValue{
					Value: []azdogit.GitRef{
						{
							Name:     stringPtr("refs/heads/main"),
							ObjectId: stringPtr("def456"),
						},
					},
				}, nil
			},
		}, "repository-id", 42)

		require.NoError(t, err)
		assert.False(t, matches)
	})
}

func TestRetryUntilPullRequestHeadMatchesRemoteBranchHead(t *testing.T) {
	originalSleep := cleanupHeadMatchSleep
	cleanupHeadMatchSleep = func(time.Duration) {}
	defer func() {
		cleanupHeadMatchSleep = originalSleep
	}()

	t.Run("retries until head matches", func(t *testing.T) {
		pr := AzureDevOps{
			Project:      "project",
			SourceBranch: "main",
		}
		attempts := 0

		matches, err := pr.retryUntilPullRequestHeadMatchesRemoteBranchHead(context.Background(), mockGitClient{
			getPullRequestIterationsFunc: func(ctx context.Context, args azdogit.GetPullRequestIterationsArgs) (*[]azdogit.GitPullRequestIteration, error) {
				attempts++
				commitID := "abc123"
				if attempts < 3 {
					commitID = "stale"
				}
				return &[]azdogit.GitPullRequestIteration{
					{
						Id: intPtr(1),
						SourceRefCommit: &azdogit.GitCommitRef{
							CommitId: &commitID,
						},
					},
				}, nil
			},
			getRefsFunc: func(ctx context.Context, args azdogit.GetRefsArgs) (*azdogit.GetRefsResponseValue, error) {
				return &azdogit.GetRefsResponseValue{
					Value: []azdogit.GitRef{
						{
							Name:     stringPtr("refs/heads/main"),
							ObjectId: stringPtr("abc123"),
						},
					},
				}, nil
			},
		}, "repository-id", 42, 3)

		require.NoError(t, err)
		assert.True(t, matches)
		assert.Equal(t, 3, attempts)
	})
}

func intPtr(v int) *int {
	return &v
}

func stringPtr(v string) *string {
	return &v
}
