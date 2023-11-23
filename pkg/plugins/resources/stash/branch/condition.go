package branch

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (g *Stash) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {

	if scm != nil {
		logrus.Warningf("scm not supported, ignoring")
	}

	if len(g.spec.Branch) == 0 {
		g.spec.Branch = source
	}

	branches, err := g.SearchBranches()
	if err != nil {
		return false, "", err
	}

	branch := source
	if g.spec.Branch != "" {
		branch = g.spec.Branch
	}

	if len(branches) == 0 {
		return false, fmt.Sprintf("no Bitbucket branch found for repository %s/%s", g.spec.Owner, g.spec.Repository), nil
	}

	for _, b := range branches {
		if b == g.spec.Branch {
			return true, fmt.Sprintf("Bitbucket branch %q found for repository %s/%s", branch, g.spec.Owner, g.spec.Repository), nil
		}
	}

	return false, fmt.Sprintf("no Bitbucket branch found matching %q for repository %s/%s",
		branch,
		g.spec.Owner,
		g.spec.Repository,
	), nil
}
