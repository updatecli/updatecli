package release

import (
	"context"
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func (g *Gitlab) Source(_ context.Context, workingDir string, resultSource *result.Source) error {
	versions, err := g.SearchReleases(g.spec.Age)

	if err != nil {
		/*
			Every published release is still cooling down, which is an expected state of
			the age filter rather than a failure, so the source is skipped instead.
		*/
		if errors.Is(err, age.ErrNoVersionMatchingAge) {
			resultSource.Result = result.SKIPPED
			resultSource.Description = "no GitLab release matches the age filter yet"
			return nil
		}

		return fmt.Errorf("searching GitLab releases: %w", err)
	}

	if len(versions) == 0 {
		return fmt.Errorf("no GitLab release found")
	}

	g.foundVersion, err = g.spec.VersionFilter.Search(versions)

	if err != nil {
		switch err {
		case version.ErrNoVersionFound:
			return fmt.Errorf("no GitLab Release tag found matching pattern %q of kind %q",
				g.versionFilter.Pattern,
				g.versionFilter.Kind,
			)
		default:
			return fmt.Errorf("filtering GitLab release: %w", err)
		}
	}

	resultSource.Information = g.foundVersion.GetVersion()

	if len(resultSource.Information) == 0 {
		return fmt.Errorf("no GitLab Release tag found matching pattern %q of kind %q",
			g.versionFilter.Pattern,
			g.versionFilter.Kind,
		)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("GitLab release tag %q found matching pattern %q of kind %q",
		resultSource.Information,
		g.versionFilter.Pattern,
		g.versionFilter.Kind,
	)
	return nil

}
