package gitcommit

import (
	"context"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest commit hash for the configured Git branch.
func (gc *GitCommit) Source(_ context.Context, workingDir string, resultSource *result.Source) error {
	var err error

	gc.directory = workingDir
	if gc.spec.URL != "" {
		gc.directory, err = gc.clone()
		if err != nil {
			return fmt.Errorf("cloning Git repository: %w", err)
		}
	} else if gc.spec.Path != "" {
		gc.directory = gc.spec.Path
	}

	if gc.directory == "" {
		return fmt.Errorf("unknown Git working directory. Did you specify one of `spec.URL`, `scmid` or `spec.path`?")
	}

	hash, err := gc.nativeGitHandler.GetCommitHash(gc.directory, gc.spec.Branch)
	if err != nil {
		return fmt.Errorf("retrieving latest Git commit: %w", err)
	}

	branch := gc.spec.Branch
	if branch == "" {
		branch = "HEAD"
	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = hash
	resultSource.Description = fmt.Sprintf("Git commit %q found for branch %q", hash, branch)

	return nil
}
