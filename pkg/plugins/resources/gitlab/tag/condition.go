package tag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (g *Gitlab) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	if scm != nil {
		logrus.Warningf("Condition not supported for the plugin GitHub Release")
	}

	tags, err := g.SearchTags()
	if err != nil {
		return false, "", fmt.Errorf("looking for GitLab tags: %w", err)
	}

	if len(tags) == 0 {
		return false, "no GitLab tag found", nil
	}

	tag := source
	if g.spec.Tag != "" {
		tag = g.spec.Tag
	}
	for _, t := range tags {
		if t == tag {
			return true, fmt.Sprintf("GitLab tag %q found", t), nil
		}
	}

	return false, fmt.Sprintf("no GitLab tag found matching pattern %q of kind %q", g.spec.VersionFilter.Pattern, g.spec.VersionFilter.Kind), nil
}
