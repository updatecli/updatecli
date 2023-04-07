package gomodule

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
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

	g := github.Github{
		Spec: github.Spec{
			URL:        "https://api.github.com",
			Owner:      parsedModule[1],
			Repository: parsedModule[2],
		},
	}

	result, err := g.ChangelogV3(version)
	if err != nil {
		logrus.Debugln(err)
	}

	return result

}
