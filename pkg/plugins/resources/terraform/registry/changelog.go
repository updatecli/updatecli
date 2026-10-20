package registry

import (
	"strings"

	"github.com/updatecli/updatecli/pkg/core/log"

	"github.com/updatecli/updatecli/pkg/core/result"
	githubChangelog "github.com/updatecli/updatecli/pkg/plugins/changelog/github/v3"
)

func (t *TerraformRegistry) Changelog(from, to string) *result.Changelogs {

	if strings.HasPrefix(t.scm, "https://github.com") {
		return getChangelogFromGitHub(t.scm, from, to)
	}
	return nil
}

func getChangelogFromGitHub(registry, from, to string) *result.Changelogs {

	splitURL := strings.Split(registry, "/")

	if len(splitURL) < 3 {
		return nil
	}

	changelog := githubChangelog.Changelog{
		Owner:      splitURL[len(splitURL)-2],
		Repository: splitURL[len(splitURL)-1],
	}

	releases, err := changelog.Search(from, to)
	if err != nil {
		log.Debugf("ignored error, searching changelogs: %s", err)
	}

	if len(releases) == 0 {
		log.Debugf("No changelog found")
		return nil
	}

	return &releases
}
