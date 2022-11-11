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
	SearchReleases(releaseType ReleaseType) (releases []string, err error)
	SearchTags() (tags []string, err error)
	Changelog(version.Version) (string, error)
}
