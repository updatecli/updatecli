package branch

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

	if len(g.spec.Branch) == 0 {
		g.spec.Branch = source
	}

	branches, err := g.SearchBranches()
	if err != nil {
		return err
	}

	branch := source
	if g.spec.Branch != "" {
		branch = g.spec.Branch
	}

	if len(branches) == 0 {

		resultCondition.Result = result.FAILURE
		resultCondition.Pass = false
		resultCondition.Description = fmt.Sprintf("no Bitbucket branch found for repository %s/%s", g.spec.Owner, g.spec.Repository)

		return nil
	}

	for _, b := range branches {
		if b == g.spec.Branch {
			resultCondition.Result = result.SUCCESS
			resultCondition.Pass = true
			resultCondition.Description = fmt.Sprintf("Bitbucket branch %q found for repository %s/%s", branch, g.spec.Owner, g.spec.Repository)

			return nil
		}
	}

	resultCondition.Result = result.FAILURE
	resultCondition.Pass = false
	resultCondition.Description = fmt.Sprintf("no Bitbucket branch found matching %q for repository %s/%s",
		branch,
		g.spec.Owner,
		g.spec.Repository,
	)

	return nil
}
