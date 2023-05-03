package gitbranch

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest git tag based on create time
func (gt *GitBranch) Source(workingDir string, resultSource *result.Source) error {

	if len(gt.spec.Path) == 0 && len(workingDir) > 0 {
		gt.spec.Path = workingDir
	}

	err := gt.Validate()
	if err != nil {
		return fmt.Errorf("validating git branch: %w", err)
	}

	tags, err := gt.nativeGitHandler.Branches(workingDir)

	if err != nil {
		return fmt.Errorf("retrieving branches: %w", err)
	}

	gt.foundVersion, err = gt.versionFilter.Search(tags)
	if err != nil {
		return fmt.Errorf("filtering branches: %w", err)
	}
	value := gt.foundVersion.GetVersion()

	if len(value) == 0 {
		return fmt.Errorf("no Git Branch found matching pattern %q", gt.versionFilter.Pattern)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = value
	resultSource.Description = fmt.Sprintf("git branch %q found matching pattern %q", value, gt.versionFilter.Pattern)

	return nil
}
