package branch

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (g *Gitea) Condition(source string) (bool, error) {
	if len(g.spec.Branch) == 0 {
		g.spec.Branch = source
	}

	branches, err := g.SearchBranches()

	if len(g.spec.Branch) == 0 {
		g.spec.Branch = source
	}

	if err != nil {
		logrus.Error(err)
		return false, err
	}

	if len(branches) == 0 {
		logrus.Infof("%s No Gitea branch found.", result.ATTENTION)
		return false, nil
	}

	for _, branch := range branches {
		if branch == g.spec.Branch {
			logrus.Infof("%s Gitea branch %q found", result.SUCCESS, branch)
			return true, nil
		}
	}

	logrus.Infof("%s No Gitea branch found matching pattern %q", result.FAILURE, g.versionFilter.Pattern)
	return false, nil
}

func (g *Gitea) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return false, fmt.Errorf("Condition not supported for the plugin GitHub Release")
}
