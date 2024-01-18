package release

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (g *Gitea) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {

	if scm != nil {
		logrus.Warningf("Condition not supported for the plugin Gitea Release")
	}

	releases, err := g.SearchReleases()
	if err != nil {
		return false, "", fmt.Errorf("looking for Gitea release: %w", err)
	}

	release := source
	if g.spec.Tag != "" {
		release = g.spec.Tag
	}

	if len(releases) == 0 {
		return false, "", fmt.Errorf("no Gitea release found")
	}

	for _, r := range releases {
		if r == release {
			return true, fmt.Sprintf("Gitea release %q found", release), nil
		}
	}

	return false, fmt.Sprintf("no Gitea Release tag found matching pattern %q", g.versionFilter.Pattern), nil
}
