package branch

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (g *Gitea) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	if scm != nil {
		logrus.Warningf("Condition not supported for the plugin GitHub Release")
	}

	branch := source
	if g.spec.Branch != "" {
		branch = g.spec.Branch
	}

	branches, err := g.SearchBranches()

	if err != nil {
		return false, "", err
	}

	if len(branches) == 0 {
		return false, "", fmt.Errorf("no Gitea branch found")
	}

	for _, b := range branches {
		if b == branch {
			return true, fmt.Sprintf("Gitea branch %q found", b), nil
		}
	}

	return false, fmt.Sprintf("no Gitea branch found matching pattern %q", g.versionFilter.Pattern), nil
}
