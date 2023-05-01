package tag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (g *Gitlab) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {
	if scm != nil {
		logrus.Warningf("Condition not supported for the plugin GitHub Release")
	}

	tags, err := g.SearchTags()
	if err != nil {
		return fmt.Errorf("looking for Gitlab tags: %w", err)
	}

	if len(tags) == 0 {
		return fmt.Errorf("no Gitlab tag found")
	}

	tag := source
	if g.spec.Tag != "" {
		tag = g.spec.Tag
	}
	for _, t := range tags {
		if t == tag {
			resultCondition.Result = result.SUCCESS
			resultCondition.Pass = true
			resultCondition.Description = fmt.Sprintf("Gitlab tag %q found", t)
			return nil
		}
	}

	return fmt.Errorf("no Gitlab tag found matching pattern %q of kind %q", g.spec.VersionFilter.Pattern, g.spec.VersionFilter.Kind)
}
