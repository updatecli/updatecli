package githubrelease

import (
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	githubChangelog "github.com/updatecli/updatecli/pkg/plugins/changelog/github/v3"
)

// Changelog returns the content (body) of the GitHub Release
func (gr GitHubRelease) Changelog(from, to string) *result.Changelogs {
	changelog := githubChangelog.Changelog{
		URL:           gr.spec.URL,
		Owner:         gr.spec.Owner,
		Repository:    gr.spec.Repository,
		Token:         gr.spec.Token,
		VersionFilter: gr.spec.VersionFilter,
	}

	releases, err := changelog.Search(from, to)
	if err != nil {
		logrus.Debugf("ignored error, searching releases: %s", err)
	}
	return &releases
}
