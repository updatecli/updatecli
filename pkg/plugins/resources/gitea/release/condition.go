package release

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (g *Gitea) Condition(source string) (bool, error) {
	releases, err := g.SearchReleases()

	if len(g.Spec.Tag) == 0 {
		g.Spec.Tag = source
	}

	if err != nil {
		logrus.Error(err)
		return false, err
	}

	if len(releases) == 0 {
		logrus.Infof("%s No Gitea release found. As a fallback you may be looking for git tags", result.ATTENTION)
		return false, fmt.Errorf("no release found, exiting")
	}

	for _, release := range releases {
		if release == g.Spec.Tag {
			logrus.Infof("%s Gitea Release tag %q found", result.SUCCESS, release)
			return true, nil
		}
	}

	logrus.Infof("%s No Gitea Release tag found matching pattern %q", result.FAILURE, g.versionFilter.Pattern)
	return false, nil

}

func (g *Gitea) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return false, fmt.Errorf("Condition not supported for the plugin Gitea Release")
}
