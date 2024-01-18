package release

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func (g *Gitea) Source(workingDir string, resultSource *result.Source) error {
	versions, err := g.SearchReleases()

	if err != nil {
		return fmt.Errorf("search gitea release: %w", err)
	}

	if len(versions) == 0 {
		return fmt.Errorf("no Gitea Release found")
	}

	g.foundVersion, err = g.spec.VersionFilter.Search(versions)

	if err != nil {
		switch err {
		case version.ErrNoVersionFound:
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
