package release

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (g *Gitlab) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {
	if scm != nil {
		logrus.Warningf("Condition not supported for the plugin Gitlab release")
	}

	releases, err := g.SearchReleases()
	if err != nil {
		return fmt.Errorf("looking for Gitlab release: %w", err)
	}

	if len(releases) == 0 {
		resultCondition.Result = result.FAILURE
		resultCondition.Pass = false
		resultCondition.Description = "no Gitlab release found"

		return nil
	}

	release := source
	if g.spec.Tag != "" {
		release = g.spec.Tag
	}
	for _, r := range releases {
		if r == release {
			resultCondition.Result = result.SUCCESS
			resultCondition.Pass = true
			resultCondition.Description = fmt.Sprintf("Gitlab release tag %q found", release)
			return nil
		}
	}

	resultCondition.Result = result.FAILURE
	resultCondition.Pass = false
	resultCondition.Description = fmt.Sprintf("no Gitlab release tag found matching pattern %q of kind %q", g.versionFilter.Pattern, g.versionFilter.Kind)

	return nil
}
