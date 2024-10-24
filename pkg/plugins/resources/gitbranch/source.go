package gitbranch

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest git tag based on create time
func (gb *GitBranch) Source(workingDir string, resultSource *result.Source) error {
	var err error

	gb.directory = workingDir
	if gb.spec.URL != "" {
		gb.directory, err = gb.clone()
		if err != nil {
			return err
		}

	} else if gb.spec.Path != "" {
		gb.directory = gb.spec.Path
	}

	err = gb.Validate()
	if err != nil {
		return fmt.Errorf("validating git branch: %w", err)
	}

	tags, err := gb.nativeGitHandler.Branches(gb.directory)

	if err != nil {
		return fmt.Errorf("retrieving branches: %w", err)
	}

	gb.foundVersion, err = gb.versionFilter.Search(tags)
	if err != nil {
		return fmt.Errorf("filtering branches: %w", err)
	}
	value := gb.foundVersion.GetVersion()

	if len(value) == 0 {
		return fmt.Errorf("no Git Branch found matching pattern %q", gb.versionFilter.Pattern)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = value
	resultSource.Description = fmt.Sprintf("git branch %q found matching pattern %q", value, gb.versionFilter.Pattern)

	return nil
}
