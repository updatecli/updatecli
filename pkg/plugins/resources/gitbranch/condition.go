package gitbranch

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks that a git branch exists
func (gt *GitBranch) Condition(source string) error {
	return gt.condition(source, nil, &result.Condition{})
}

// ConditionFromSCM test if a branch exists from a git repository specific from SCM
func (gt *GitBranch) ConditionFromSCM(source string, scm scm.ScmHandler) error {

	return gt.condition(source, nil, &result.Condition{})
}

func (gt *GitBranch) condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {

	if scm != nil {
		path := scm.GetDirectory()

		if len(gt.spec.Path) > 0 {
			logrus.Warningf("Path is defined and set to %q but is overridden by the scm definition %q",
				gt.spec.Path,
				path)
		}
		gt.spec.Path = path
	}

	gt.branch = gt.spec.Branch
	// If source input is empty, then it means that it was disabled by the user with `disablesourceinput: true`
	if source != "" {
		logrus.Infof("Source input value detected")
		gt.branch = source
	}

	err := gt.Validate()
	if err != nil {
		return fmt.Errorf("git tag validation: %w", err)
	}

	branches, err := gt.nativeGitHandler.Branches(gt.spec.Path)
	if err != nil {
		return fmt.Errorf("searching git branches: %w", err)
	}

	for _, b := range branches {
		if b == gt.branch {
			resultCondition.Pass = true
			resultCondition.Result = result.SUCCESS
			resultCondition.Description = fmt.Sprintf("git branch %q matching", gt.branch)
			return nil
		}
	}

	resultCondition.Pass = false
	resultCondition.Result = result.FAILURE
	resultCondition.Description = fmt.Sprintf("git branch %q not found", gt.branch)

	return nil
}
