package tag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (g *Gitea) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {
	if scm != nil {
		logrus.Warningf("Condition not supported for the plugin Gitea tag")
	}

	tag := source
	if g.spec.Tag != "" {
		tag = g.spec.Tag
	}

	tags, err := g.SearchTags()
	if err != nil {
		return fmt.Errorf("looking for Gitea tag: %w", err)
	}

	if len(tags) == 0 {
		return fmt.Errorf("no Gitea Tags found")
	}

	for _, t := range tags {
		if t == tag {
			resultCondition.Pass = true
			resultCondition.Result = result.SUCCESS
			resultCondition.Description = fmt.Sprintf("Gitea tag %q found", t)
			return nil
		}
	}

	resultCondition.Description = fmt.Sprintf("no Gitea tag found matching %q", g.spec.Tag)
	resultCondition.Pass = false
	resultCondition.Result = result.FAILURE

	return nil
}
