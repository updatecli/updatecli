package github

import (
	"context"

	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// GithubHandler must be implemented by any GitHub module
type GithubHandler interface {
	SearchReleases(ctx context.Context, releaseType ReleaseType, retry int) (releases []ReleaseNode, err error)
	SearchReleasesByTagName(ctx context.Context, releaseType ReleaseType) (releases []string, err error)
	SearchReleasesByTagHash(ctx context.Context, releaseType ReleaseType) (releases []string, err error)
	SearchReleasesByTitle(ctx context.Context, releaseType ReleaseType) (releases []string, err error)
	SearchTags(ctx context.Context, retry int) (tags []string, err error)
	Changelog(version.Version) (string, error)
}
