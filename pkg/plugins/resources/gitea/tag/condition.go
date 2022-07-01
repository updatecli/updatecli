package tag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (g *Gitea) Condition(source string) (bool, error) {

	if len(g.Spec.Tag) == 0 {
		g.Spec.Tag = source
	}

	tags, err := g.SearchTags()

	if err != nil {
		logrus.Error(err)
		return false, err
	}

	if len(tags) == 0 {
		logrus.Infof("%s No Gitea Tags found.", result.ATTENTION)
		if len(tags) == 0 {
			logrus.Infof("\t=> No Gitea tags found, exiting")
			return false, fmt.Errorf("no Gitea tags found, exiting")
		}
	}

	for _, tag := range tags {
		if tag == g.Spec.Tag {
			logrus.Infof("%s Gitea tag %q found", result.SUCCESS, tag)
			return true, nil
		}
	}

	logrus.Infof("%s No Gitea Tags found matching pattern %q", result.FAILURE, g.versionFilter.Pattern)
	return false, nil

}

func (g *Gitea) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return false, fmt.Errorf("Condition not supported for the plugin GitHub Release")
}
