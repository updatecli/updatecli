package gitbranch

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition checks that a git branch exists
func (gt *GitBranch) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {

	if gt.spec.Path != "" && scm != nil {
		logrus.Warningf("Path setting value %q is overriding the scm configuration (value %q)",
			gt.spec.Path,
			scm.GetDirectory())
	}

	if gt.spec.URL != "" && scm != nil {
		logrus.Warningf("URL setting value %q is overriding the scm configuration (value %q)",
			gt.spec.URL,
			scm.GetURL())
	}

	if gt.spec.URL != "" {
		gt.directory, err = gt.clone()
		if err != nil {
			return false, "", err
		}

	} else if gt.spec.Path != "" {
		gt.directory = gt.spec.Path
	} else if scm != nil {
		gt.directory = scm.GetDirectory()
	}

	gt.branch = source
	// If source input is empty, then it means that it was disabled by the user with `disablesourceinput: true`
	if gt.spec.Branch != "" {
		gt.branch = gt.spec.Branch
	}

	err = gt.Validate()
	if err != nil {
		return false, "", fmt.Errorf("git tag validation: %w", err)
	}

	branches, err := gt.nativeGitHandler.Branches(gt.directory)
	if err != nil {
		return false, "", fmt.Errorf("searching git branches: %w", err)
	}

	for _, b := range branches {
		if b == gt.branch {
			return true, fmt.Sprintf("git branch %q matching", gt.branch), nil
		}
	}

	return false, fmt.Sprintf("git branch %q not found", gt.branch), nil
}
