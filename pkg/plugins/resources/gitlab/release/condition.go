package release

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (g *Gitlab) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	if scm != nil {
		logrus.Warningf("Condition not supported for the plugin GitLab release")
	}

	releases, err := g.SearchReleases()
	if err != nil {
		return false, "", fmt.Errorf("looking for GitLab release: %w", err)
	}

	if len(releases) == 0 {
		return false, "no GitLab release found", nil
	}

	release := source
	if g.spec.Tag != "" {
		release = g.spec.Tag
	}
	for _, r := range releases {
		if r == release {

			return true, fmt.Sprintf("GitLab release tag %q found", release), nil
		}
	}

	return false, fmt.Sprintf("no GitLab release tag found matching pattern %q of kind %q", g.versionFilter.Pattern, g.versionFilter.Kind), nil
}
