package tag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (g *Gitea) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	if scm != nil {
		logrus.Warningf("Condition not supported for the plugin Gitea tag")
	}

	tag := source
	if g.spec.Tag != "" {
		tag = g.spec.Tag
	}

	tags, err := g.SearchTags()
	if err != nil {
		return false, "", fmt.Errorf("looking for Gitea tag: %w", err)
	}

	if len(tags) == 0 {
		return false, "", fmt.Errorf("no Gitea Tags found")
	}

	for _, t := range tags {
		if t == tag {
			return true, fmt.Sprintf("Gitea tag %q found", t), nil
		}
	}

	return false, fmt.Sprintf("no Gitea tag found matching %q", g.spec.Tag), nil
}
