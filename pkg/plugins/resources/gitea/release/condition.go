package release

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (g *Gitea) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {

	if scm != nil {
		logrus.Warningf("Condition not supported for the plugin Gitea Release")
	}

	releases, err := g.SearchReleases()
	if err != nil {
		return fmt.Errorf("looking for Gitea release: %w", err)
	}

	release := source
	if g.spec.Tag != "" {
		release = g.spec.Tag
	}

	if len(releases) == 0 {
		return fmt.Errorf("no Gitea release found")
	}

	for _, r := range releases {
		if r == release {
			resultCondition.Pass = true
			resultCondition.Description = fmt.Sprintf("Gitea release %q found", release)
			resultCondition.Result = result.SUCCESS

			return nil
		}
	}

	return fmt.Errorf("no Gitea Release tag found matching pattern %q", g.versionFilter.Pattern)
}
