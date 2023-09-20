package registry

import (
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
)

func (t *TerraformRegistry) Changelog() string {
	if strings.HasPrefix(t.scm, "https://github.com") {
		splitURL := strings.Split(t.scm, "/")
		return getChangelogFromGitHub(splitURL[len(splitURL)-2], splitURL[len(splitURL)-1], t.Version.OriginalVersion)
	}
	return ""
}

func getChangelogFromGitHub(owner, repo, version string) string {
	g := github.Github{
		Spec: github.Spec{
			URL:        "https://api.github.com",
			Owner:      owner,
			Repository: repo,
		},
	}

	var result string
	result, err := g.ChangelogV3(fmt.Sprintf("v%s", version))
	if err != nil {
		logrus.Debugln(err)
	}

	// Try without a v prefix
	if result == "" {
		result, err = g.ChangelogV3(version)
		if err != nil {
			logrus.Debugln(err)
		}
	}

	return result
}
