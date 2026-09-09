package branch

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
)

func (g *Gitlab) Condition(_ context.Context, source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	if scm != nil {
		return false, "", fmt.Errorf("Condition not supported for the plugin GitLab branch")
	}

	// A condition checks whether a specific branch exists, so the age filter, which
	// only narrows down which branch to pick, doesn't apply here.
	branches, err := g.SearchBranches(age.Spec{})
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
