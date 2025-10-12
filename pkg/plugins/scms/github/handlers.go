package github

import "github.com/updatecli/updatecli/pkg/plugins/utils/version"

// GithubHandler must be implemented by any GitHub module
type GithubHandler interface {
	SearchReleases(releaseType ReleaseType, retry int) (releases []ReleaseNode, err error)
	SearchReleasesByTagName(releaseType ReleaseType) (releases []string, err error)
	SearchReleasesByTagHash(releaseType ReleaseType) (releases []string, err error)
	SearchReleasesByTitle(releaseType ReleaseType) (releases []string, err error)
	SearchTags(retry int) (tags []string, err error)
	Changelog(version.Version) (string, error)
}
