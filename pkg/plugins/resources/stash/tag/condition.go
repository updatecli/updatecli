package tag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (g *Stash) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {

	if scm != nil {
		logrus.Warningf("scm not supported, ignored")
	}

	tag := source
	if g.spec.Tag != "" {
		tag = g.spec.Tag
	}

	tags, err := g.SearchTags()
	if err != nil {
		return false, "", fmt.Errorf("looking for tag: %w", err)
	}

	if len(tags) == 0 {
		return false, fmt.Sprintf("no Bitbucket Tags found for %s/%s", g.spec.Owner, g.spec.Repository), nil
	}

	for _, t := range tags {
		if t == g.spec.Tag {
			return true, fmt.Sprintf("bitbucket tag %q found", tag), nil
		}
	}

	return false, fmt.Sprintf("no Bitbucket Tags found matching %q", tag), nil
}
