package release

import (
	"context"
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func (g *Gitea) Source(ctx context.Context, pathResolver pathresolver.Resolver, resultSource *result.Source) error {
	versions, published, err := g.searchReleases(ctx, g.spec.Age)
	if err != nil {
		return fmt.Errorf("search gitea release: %w", err)
	}

	if len(published) == 0 {
		return fmt.Errorf("no Gitea Release found")
	}

	g.foundVersion, err = age.Search(&g.spec.VersionFilter, published, versions)
	if err != nil {
		/*
			Every release matching the version filter is still cooling down, which is an
			expected state of the age filter rather than a failure, so the source is skipped
			instead.
		*/
		if errors.Is(err, age.ErrNoVersionMatchingAge) {
			resultSource.Result = result.SKIPPED
			resultSource.Description = "no Gitea Release matching the version filter matches the age filter yet"
			return nil
		}

		switch {
		case errors.Is(err, version.ErrNoVersionFound):
			return fmt.Errorf("no Gitea Release tag found matching pattern %q", g.versionFilter.Pattern)
		default:
			return fmt.Errorf("searching version matching pattern: %w", err)
		}
	}

	value := g.foundVersion.GetVersion()

	if len(value) == 0 {
		return fmt.Errorf("no Gitea Release tag found matching pattern %q", g.versionFilter.Pattern)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = value
	resultSource.Description = fmt.Sprintf("Gitea Release tag %q found matching pattern %q of kind %q", value, g.versionFilter.Pattern, g.versionFilter.Kind)

	return nil
}
