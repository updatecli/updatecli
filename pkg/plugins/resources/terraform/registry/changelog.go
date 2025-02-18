package registry

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"

	githubChangelog "github.com/updatecli/updatecli/pkg/plugins/changelog/github/v3"
)

func (t *TerraformRegistry) Changelog() string {
	if strings.HasPrefix(t.scm, "https://github.com") {
		splitURL := strings.Split(t.scm, "/")
		return getChangelogFromGitHub(splitURL[len(splitURL)-2], splitURL[len(splitURL)-1], t.Version.OriginalVersion)
	}
	return ""
}

func getChangelogFromGitHub(owner, repo, version string) string {

	changelog := githubChangelog.Changelog{
		Owner:      owner,
		Repository: repo,
	}

	releases, err := changelog.Search(version, version)

	if err != nil {
		logrus.Debugf("ignored error, searching changelogs: %s", err)
	}

	if len(releases) == 0 {
		if releases, err = changelog.Search(version, version); err != nil {
			logrus.Debugf("ignored error, searching for changelogs: %s", err)
		}
	}

	if len(releases) == 0 {
		return ""
	}

	return fmt.Sprintf("\nRelease published on the %v at the url %v\n\n%v",
		releases[0].PublishedAt,
		releases[0].URL,
		releases[0].Body)
}
