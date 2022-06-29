package release

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (g *Gitea) Condition(source string) (bool, error) {
	versions, err := g.SearchReleases()

	if err != nil {
		logrus.Error(err)
		return false, err
	}

	if len(versions) == 0 {
		logrus.Infof("%s No Gitea Release found. As a fallback you may be looking for git tags", result.ATTENTION)
		if len(versions) == 0 {
			logrus.Infof("\t=> No release found, exiting")
			return false, fmt.Errorf("no release found, exiting")
		}
	}

	g.foundVersion, err = g.versionFilter.Search(versions)
	if err != nil {
		return false, err
	}
	value := g.foundVersion.ParsedVersion

	if len(value) == 0 {
		logrus.Infof("%s No Gitea Release version found matching pattern %q", result.FAILURE, g.versionFilter.Pattern)
		return false, nil
	} else if len(value) > 0 {
		logrus.Infof("%s Gitea Release version %q found matching pattern %q", result.SUCCESS, value, g.versionFilter.Pattern)
		return true, nil
	}

	err = fmt.Errorf("something unexpected happened in Github source")
	return false, err

}

func (g *Gitea) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return false, fmt.Errorf("Condition not supported for the plugin GitHub Release")
}
