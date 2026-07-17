package gitcommit

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition checks that a Git commit exists in the repository.
func (gc *GitCommit) Condition(_ context.Context, source string, scm scm.ScmHandler) (pass bool, message string, err error) {

	if gc.spec.Path != "" && scm != nil {
		logrus.Warningf("Path setting value %q is overriding the scm configuration (value %q)",
			gc.spec.Path,
			scm.GetDirectory())
	}

	if gc.spec.URL != "" && scm != nil {
		logrus.Warningf("URL setting value %q is overriding the scm configuration (value %q)",
			gc.spec.URL,
			scm.GetURL())
	}

	if gc.spec.URL != "" {
		gc.directory, err = gc.clone()
		if err != nil {
			return false, "", fmt.Errorf("cloning Git repository: %w", err)
		}
	} else if gc.spec.Path != "" {
		gc.directory = gc.spec.Path
	} else if scm != nil {
		gc.directory = scm.GetDirectory()
	}

	if gc.directory == "" {
		return false, "", fmt.Errorf("unknown Git working directory. Did you specify one of `spec.URL`, `scmid` or `spec.path`?")
	}

	commit := source
	// If source input is empty, then it means that it was disabled by the user with `disablesourceinput: true`
	if gc.spec.Hash != "" {
		commit = gc.spec.Hash
	}

	if commit == "" {
		return false, "", fmt.Errorf("unknown Git commit. Did you specify a source input or `spec.hash`?")
	}

	found, err := gc.nativeGitHandler.IsCommitExist(gc.directory, commit)
	if err != nil {
		return false, "", fmt.Errorf("checking Git commit existence: %w", err)
	}

	if found {
		return true, fmt.Sprintf("git commit %q found", commit), nil
	}

	return false, fmt.Sprintf("git commit %q not found", commit), nil
}
