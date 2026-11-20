package release

import (
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/core/result"
	giteaChangelog "github.com/updatecli/updatecli/pkg/plugins/changelog/gitea"
)

// Changelog returns release notes between two versions, including both endpoints.
// For version kinds without range support, it returns only the new release.
func (g *Gitea) Changelog(from, to string) *result.Changelogs {
	changelog := giteaChangelog.Changelog{
		URL:           g.spec.URL,
		Owner:         g.spec.Owner,
		Repository:    g.spec.Repository,
		Token:         g.spec.Token,
		VersionFilter: g.versionFilter,
	}

	changelogs, err := changelog.Search(from, to)
	if err != nil {
		log.Debugf("failed to retrieve Gitea changelog: %s", err)
		return nil
	}
	if len(changelogs) == 0 {
		return nil
	}

	return &changelogs
}
