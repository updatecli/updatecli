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
		return fmt.Errorf("looking for GitLab tags: %w", err)
	}

	if len(tags) == 0 {
		resultCondition.Result = result.FAILURE
		resultCondition.Pass = false
		resultCondition.Description = "no GitLab tag found"
		return nil
	}

	tag := source
	if g.spec.Tag != "" {
		tag = g.spec.Tag
	}
	for _, t := range tags {
		if t == tag {
			resultCondition.Result = result.SUCCESS
			resultCondition.Pass = true
			resultCondition.Description = fmt.Sprintf("GitLab tag %q found", t)
			return nil
		}
	}

	resultCondition.Result = result.FAILURE
	resultCondition.Pass = false
	resultCondition.Description = fmt.Sprintf("no GitLab tag found matching pattern %q of kind %q", g.spec.VersionFilter.Pattern, g.spec.VersionFilter.Kind)

	return nil
}
