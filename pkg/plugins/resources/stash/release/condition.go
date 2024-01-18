package release

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (g *Stash) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {

	if scm != nil {
		logrus.Warningf("scm not supported, ignoring")
	}

	releases, err := g.SearchReleases()
	if err != nil {
		return fmt.Errorf("looking for releases: %w", err)
	}

	release := source
	if g.spec.Tag != "" {
		release = g.spec.Tag
	}

	if len(releases) == 0 {
		resultCondition.Result = result.FAILURE
		resultCondition.Pass = false
		resultCondition.Description = fmt.Sprintf("no Bitbucket release found for repository %s/%s", g.spec.Owner, g.spec.Repository)
		return nil
	}

	for _, r := range releases {
		if r == g.spec.Tag {
			resultCondition.Result = result.SUCCESS
			resultCondition.Pass = true
			resultCondition.Description = fmt.Sprintf("bitbucket release tag %q found", release)
			return nil
		}
	}

	resultCondition.Result = result.FAILURE
	resultCondition.Pass = false
	resultCondition.Description = fmt.Sprintf("no Bitbucket Release tag found matching %q for repository %s%s",
		release,
		g.spec.Owner,
		g.spec.Repository,
	)

	return nil

}
