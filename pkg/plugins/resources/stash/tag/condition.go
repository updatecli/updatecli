package tag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (g *Stash) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {

	if scm != nil {
		logrus.Warningf("scm not supported, ignored")
	}

	tag := source
	if g.spec.Tag != "" {
		tag = g.spec.Tag
	}

	tags, err := g.SearchTags()
	if err != nil {
		return fmt.Errorf("looking for tag: %w", err)
	}

	if len(tags) == 0 {
		resultCondition.Description = fmt.Sprintf("no Bitbucket Tags found for %s/%s", g.spec.Owner, g.spec.Repository)
		resultCondition.Pass = false
		resultCondition.Result = result.FAILURE
		return nil
	}

	for _, t := range tags {
		if t == g.spec.Tag {
			resultCondition.Pass = true
			resultCondition.Result = result.SUCCESS
			resultCondition.Description = fmt.Sprintf("bitbucket tag %q found", tag)
			return nil
		}
	}

	resultCondition.Description = fmt.Sprintf("no Bitbucket Tags found matching %q", tag)
	resultCondition.Pass = false
	resultCondition.Result = result.FAILURE

	return nil

}
