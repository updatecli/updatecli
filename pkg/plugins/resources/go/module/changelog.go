package gomodule

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	githubChangelog "github.com/updatecli/updatecli/pkg/plugins/changelog/github/v3"
)

// Changelog returns the changelog for a specific golang module, or an empty string if it couldn't find one
func (g *GoModule) Changelog(from, to string) *result.Changelogs {

	if from == "" && to == "" {
		return nil
	}

	if strings.HasPrefix(g.Spec.Module, "github.com") {
		return getChangelogFromGitHub(g.Spec.Module, from, to)
	}

	return nil
}

func getChangelogFromGitHub(module, from, to string) *result.Changelogs {
	parsedModule := strings.Split(module, "/")

	if len(parsedModule) < 3 {
		return nil
	}

	changelog := githubChangelog.Changelog{
		Owner:      parsedModule[1],
		Repository: parsedModule[2],
	}

	releases, err := changelog.Search(from, to)
	if err != nil {
		logrus.Debugf("ignored error, searching releases: %s", err)
	}

	return &releases
}
