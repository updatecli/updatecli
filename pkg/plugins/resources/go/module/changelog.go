package gomodule

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	githubChangelog "github.com/updatecli/updatecli/pkg/plugins/changelog/github/v3"
)

// Changelog returns the changelog for a specific golang module, or an empty string if it couldn't find one
func (g *GoModule) Changelog() string {
	if strings.HasPrefix(g.Spec.Module, "github.com") {
		return getChangelogFromGitHub(g.Spec.Module, g.Version.OriginalVersion)
	}

	return ""
}

func getChangelogFromGitHub(module, version string) string {
	parsedModule := strings.Split(module, "/")

	if len(parsedModule) < 3 {
		return ""
	}

	changelog := githubChangelog.Changelog{
		Owner:      parsedModule[1],
		Repository: parsedModule[2],
	}

	releases, err := changelog.Search(version, version)

	if err != nil {
		logrus.Debugf("ignored error, searching releases: %s", err)
	}

	if len(releases) == 0 {
		return ""
	}

	return fmt.Sprintf("\nRelease published on the %v at the url %v\n\n%v",
		releases[0].PublishedAt,
		releases[0].URL,
		releases[0].Body)
}
