package tag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (g *Gitea) Condition(source string) (bool, error) {
	versions, err := g.SearchTags()

	if err != nil {
		logrus.Error(err)
		return false, err
	}

	if len(versions) == 0 {
		logrus.Infof("%s No Gitea Tags found.", result.ATTENTION)
		if len(versions) == 0 {
			logrus.Infof("\t=> No Gitea tags found, exiting")
			return false, fmt.Errorf("no Gitea tags found, exiting")
		}
	}

	g.foundVersion, err = g.versionFilter.Search(versions)
	if err != nil {
		return false, err
	}
	value := g.foundVersion.ParsedVersion

	if len(value) == 0 {
		logrus.Infof("%s No Gitea Tags version found matching pattern %q", result.FAILURE, g.versionFilter.Pattern)
		return false, nil
	} else if len(value) > 0 {
		logrus.Infof("%s Gitea Tags version %q found matching pattern %q", result.SUCCESS, value, g.versionFilter.Pattern)
		return true, nil
	}

	err = fmt.Errorf("something unexpected happened in Github source")
	return false, err

}

func (g *Gitea) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return false, fmt.Errorf("Condition not supported for the plugin GitHub Release")
}
