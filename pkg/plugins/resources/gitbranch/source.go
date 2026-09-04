package gitbranch

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
)

// Source returns the latest git tag based on create time
func (gb *GitBranch) Source(_ context.Context, workingDir string, resultSource *result.Source) error {
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

	if gb.directory == "" {
		return fmt.Errorf("unknown Git working directory. Did you specify one of `spec.URL`, `scmid` or a `spec.path`?")
	}

	refs, err := gb.nativeGitHandler.BranchRefs(gb.directory)

	if err != nil {
		return fmt.Errorf("retrieving branches: %w", err)
	}

	values := []string{}
	for _, ref := range refs {
		if !gb.spec.Age.Matches(ref.When) {
			logrus.Debugf("ignoring branch %q, dated %s, as outside of the age window", ref.Name, ref.When)
			continue
		}
		values = append(values, ref.Name)
	}

	/*
		The repository does publish branches but the age filter discarded every one of
		them, which is an expected state of the filter rather than a failure, so the
		source is skipped instead.
	*/
	if len(refs) > 0 && len(values) == 0 {
		logrus.Debugf("%s", age.ErrNoVersionMatchingAge)
		resultSource.Result = result.SKIPPED
		resultSource.Description = "no git branch matches the age filter yet"
		return nil
	}

	gb.foundVersion, err = gb.versionFilter.Search(values)
	if err != nil {
		return fmt.Errorf("filtering branches: %w", err)
	}
	value := gb.foundVersion.GetVersion()

	if len(value) == 0 {
		return fmt.Errorf("no Git Branch found matching pattern %q", gb.versionFilter.Pattern)
	}

	if gb.spec.Key == "hash" {
		for i := range refs {
			if refs[i].Name == value {
				value = refs[i].Hash
			}
		}
	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = value
	resultSource.Description = fmt.Sprintf("git branch %q found matching pattern %q", value, gb.versionFilter.Pattern)

	return nil
}
