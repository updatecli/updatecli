package branch

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (g *Gitea) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {
	if scm != nil {
		logrus.Warningf("Condition not supported for the plugin GitHub Release")
	}

	branch := source
	if g.spec.Branch != "" {
		branch = g.spec.Branch
	}

	branches, err := g.SearchBranches()

	if err != nil {
		return err
	}

	if len(branches) == 0 {
		return fmt.Errorf("no Gitea branch found")
	}

	for _, b := range branches {
		if b == branch {
			resultCondition.Result = result.SUCCESS
			resultCondition.Pass = true
			resultCondition.Description = fmt.Sprintf("Gitea branch %q found", b)
			return nil
		}
	}

	resultCondition.Result = result.FAILURE
	resultCondition.Description = fmt.Sprintf("no Gitea branch found matching pattern %q", g.versionFilter.Pattern)

	return nil
}
