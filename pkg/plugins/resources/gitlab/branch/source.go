package branch

import (
	"context"
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func (g *Gitlab) Source(_ context.Context, workingDir string, resultSource *result.Source) error {
	versions, err := g.SearchBranches(g.spec.Age)

	if err != nil {
		/*
			Every branch is still cooling down, which is an expected state of the age
			filter rather than a failure, so the source is skipped instead.
		*/
		if errors.Is(err, age.ErrNoVersionMatchingAge) {
			resultSource.Result = result.SKIPPED
			resultSource.Description = "no GitLab branch matches the age filter yet"
			return nil
		}

		return fmt.Errorf("searching GitLab branches: %q", err)
	}

	if len(versions) == 0 {
		return fmt.Errorf("no GitLab branches found")
	}

	g.foundVersion, err = g.spec.VersionFilter.Search(versions)

	if err != nil {
		switch err {
		case version.ErrNoVersionFound:
			return fmt.Errorf("no GitLab branches found matching pattern %q of kind %q",
				g.versionFilter.Pattern,
				g.versionFilter.Kind,
			)

		default:
			return fmt.Errorf("filtering GitLab branches matching pattern %q of kind %q",
				g.versionFilter.Pattern,
				g.versionFilter.Kind,
			)
		}
	}

	value := g.foundVersion.GetVersion()

	if len(value) == 0 {
		return fmt.Errorf("no GitLab branches found matching pattern %q of kind %q",
			g.versionFilter.Pattern,
			g.versionFilter.Kind,
		)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = value
	resultSource.Description = fmt.Sprintf("GitLab branches %q found matching pattern %q of kind %q",
		value,
		g.versionFilter.Pattern,
		g.versionFilter.Kind)

	return nil
}
