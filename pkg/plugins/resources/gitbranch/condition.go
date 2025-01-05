package gitbranch

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition checks that a git branch exists
func (gb *GitBranch) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {

	if gb.spec.Path != "" && scm != nil {
		logrus.Warningf("Path setting value %q is overriding the scm configuration (value %q)",
			gb.spec.Path,
			scm.GetDirectory())
	}

	if gb.spec.URL != "" && scm != nil {
		logrus.Warningf("URL setting value %q is overriding the scm configuration (value %q)",
			gb.spec.URL,
			scm.GetURL())
	}

	if gb.spec.URL != "" {
		gb.directory, err = gb.clone()
		if err != nil {
			return false, "", err
		}

	} else if gb.spec.Path != "" {
		gb.directory = gb.spec.Path
	} else if scm != nil {
		gb.directory = scm.GetDirectory()
	}

	gb.branch = source
	// If source input is empty, then it means that it was disabled by the user with `disablesourceinput: true`
	if gb.spec.Branch != "" {
		gb.branch = gb.spec.Branch
	}

	if gb.directory == "" {
		return false, "", fmt.Errorf("Unknown Git working directory. Did you specify one of `spec.URL`, `scmid` or a `spec.path`?")
	}

	branches, err := gb.nativeGitHandler.Branches(gb.directory)
	if err != nil {
		return false, "", fmt.Errorf("searching git branches: %w", err)
	}

	for _, b := range branches {
		if b == gb.branch {
			return true, fmt.Sprintf("git branch %q matching", gb.branch), nil
		}
	}

	return false, fmt.Sprintf("git branch %q not found", gb.branch), nil
}
