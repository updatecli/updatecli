package tag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (g *Gitlab) Condition(source string) (bool, error) {

	if len(g.spec.Tag) == 0 {
		g.spec.Tag = source
	}

	tags, err := g.SearchTags()

	if err != nil {
		logrus.Error(err)
		return false, err
	}

	if len(tags) == 0 {
		logrus.Infof("%s No Gitlab Tags found.", result.ATTENTION)
		return false, nil
	}

	for _, tag := range tags {
		if tag == g.spec.Tag {
			logrus.Infof("%s Gitlab tag %q found", result.SUCCESS, tag)
			return true, nil
		}
	}

	logrus.Infof("%s No Gitlab Tags found matching  %q", result.FAILURE, g.spec.Tag)
	return false, nil

}

func (g *Gitlab) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return false, fmt.Errorf("Condition not supported for the plugin GitHub Release")
}
