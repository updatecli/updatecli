package branch

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (g *Gitea) Condition(_ context.Context, source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	if scm != nil {
		logrus.Warningf("scm is not supported for the Gitea branch condition, ignoring")
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
