package branch

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (g *Gitlab) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {
	if scm != nil {
		return fmt.Errorf("Condition not supported for the plugin GitLab branch")
	}

	branches, err := g.SearchBranches()
	if err != nil {
		return fmt.Errorf("looking for GitLab branch: %w", err)
	}

	if len(branches) == 0 {
		resultCondition.Pass = false
		resultCondition.Result = result.FAILURE
		resultCondition.Description = "no GitLab branch found"
		return nil
	}

	branch := source
	if g.spec.Branch != "" {
		branch = g.spec.Branch
	}
	for _, b := range branches {
		if b == branch {
			resultCondition.Pass = true
			resultCondition.Result = result.SUCCESS
			resultCondition.Description = fmt.Sprintf("GitLab branch %q found", b)
			return nil
		}
	}

	resultCondition.Result = result.FAILURE
	resultCondition.Pass = false
	resultCondition.Description = fmt.Sprintf("no GitLab branch found matching pattern %q", g.versionFilter.Pattern)

	return nil
}
