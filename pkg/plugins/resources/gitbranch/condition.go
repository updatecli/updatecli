package gitbranch

import (
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks that a git branch exists
func (gt *GitBranch) Condition(source string) (bool, error) {
	return gt.condition(source)
}

// ConditionFromSCM test if a branch exists from a git repository specific from SCM
func (gt *GitBranch) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	path := scm.GetDirectory()

	if len(gt.spec.Path) > 0 {
		logrus.Warningf("Path is defined and set to %q but is overridden by the scm definition %q",
			gt.spec.Path,
			path)
	}
	gt.spec.Path = path

	return gt.condition(source)
}

func (gt *GitBranch) condition(source string) (bool, error) {

	gt.branch = gt.spec.Branch
	// If source input is empty, then it means that it was disabled by the user with `disablesourceinput: true`
	if source != "" {
		logrus.Infof("Source input value detected")
		gt.branch = source
	}

	err := gt.Validate()
	if err != nil {
		return false, err
	}

	branches, err := gt.nativeGitHandler.Branches(gt.spec.Path)
	if err != nil {
		return false, err
	}

	found := false
	for _, b := range branches {
		if b == gt.branch {
			found = true
		}
	}

	if found {
		logrus.Printf("%s Git branch %q matching\n", result.SUCCESS, gt.branch)
		return true, nil
	}

	logrus.Printf("%s Git branch %q not found\n",
		result.FAILURE,
		gt.branch)

	return false, nil
}
