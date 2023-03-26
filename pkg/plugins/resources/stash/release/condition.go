package release

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (g *Stash) Condition(source string) (bool, error) {
	releases, err := g.SearchReleases()

	if len(g.spec.Tag) == 0 {
		g.spec.Tag = source
	}

	if err != nil {
		logrus.Error(err)
		return false, err
	}

	if len(releases) == 0 {
		logrus.Infof("%s No Bitbucket release found. As a fallback you may be looking for git tags", result.ATTENTION)
		return false, nil
	}

	for _, release := range releases {
		if release == g.spec.Tag {
			logrus.Infof("%s Bitbucket Release tag %q found", result.SUCCESS, release)
			return true, nil
		}
	}

	logrus.Infof("%s No Bitbucket Release tag found matching pattern %q", result.FAILURE, g.versionFilter.Pattern)
	return false, nil

}

func (g *Stash) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return false, fmt.Errorf("Condition not supported for the plugin Bitbucket Release")
}
