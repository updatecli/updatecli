package branch

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func (g *Gitlab) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	if scm != nil {
		return false, "", fmt.Errorf("Condition not supported for the plugin GitLab branch")
	}

	branches, err := g.SearchBranches()
	if err != nil {
		return false, "", fmt.Errorf("looking for GitLab branch: %w", err)
	}

	if len(branches) == 0 {
		return false, "no GitLab branch found", nil
	}

	branch := source
	if g.spec.Branch != "" {
		branch = g.spec.Branch
	}
	for _, b := range branches {
		if b == branch {
			return true, fmt.Sprintf("GitLab branch %q found", b), nil
		}
	}

	return false, fmt.Sprintf("no GitLab branch found matching pattern %q", g.versionFilter.Pattern), nil
}
