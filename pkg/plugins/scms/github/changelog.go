package github

import (
	"fmt"

	githubChangelog "github.com/updatecli/updatecli/pkg/plugins/changelog/github/v3"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Changelog returns a changelog description based on a release name
func (g *Github) Changelog(version version.Version) (string, error) {

	// GitHub Release needs the original version, because the "found" version can be modified (semantic version without the prefix, transformed version, etc.)
	versionName := version.OriginalVersion

	accessToken, err := GetAccessToken(g.token)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	changelog := githubChangelog.Changelog{
		URL:        g.GetURL(),
		Owner:      g.Spec.Owner,
		Repository: g.Spec.Repository,
		Token:      accessToken,
	}

	releases, err := changelog.Search(versionName, versionName)

	if err != nil {
		return "", fmt.Errorf("searching for github release changelog: %w", err)
	}

	if len(releases) == 0 {
		return "", fmt.Errorf("no release detected")
	}

	return fmt.Sprintf("\nRelease published on the %v at the url %v\n\n%v",
		releases[0].PublishedAt,
		releases[0].URL,
		releases[0].Body), nil
}
