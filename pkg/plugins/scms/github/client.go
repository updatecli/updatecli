package github

import (
	"context"

	"github.com/shurcooL/githubv4"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// GitHubClient must be implemented by any GitHub query client (v4 API)
type GitHubClient interface {
	Query(ctx context.Context, q interface{}, variables map[string]interface{}) error
	Mutate(ctx context.Context, m interface{}, input githubv4.Input, variables map[string]interface{}) error
}

// GithubHandler must be implemented by any GitHub module
type GithubHandler interface {
	SearchReleases(releaseType ReleaseType, retry int) (releases []ReleaseNode, err error)
	SearchReleasesByTagName(releaseType ReleaseType) (releases []string, err error)
	SearchReleasesByTagHash(releaseType ReleaseType) (releases []string, err error)
	SearchReleasesByTitle(releaseType ReleaseType) (releases []string, err error)
	SearchTags(retry int) (tags []string, err error)
	Changelog(version.Version) (string, error)
}

const (
	ErrAPIRateLimitExceeded             = "API rate limit already exceeded"
	ErrAPIRateLimitExceededFinalAttempt = "API rate limit exceeded, final attempt failed"
)

var (
	MaxRetry = 3
)
