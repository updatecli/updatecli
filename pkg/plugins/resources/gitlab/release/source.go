package release

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func (g *Gitlab) Source(workingDir string, resultSource *result.Source) error {
	versions, err := g.SearchReleases()

	if err != nil {
		return fmt.Errorf("searching Gitlab releases: %w", err)
	}

	if len(versions) == 0 {
		return fmt.Errorf("no Gitlab release found")
	}

	g.foundVersion, err = g.spec.VersionFilter.Search(versions)

	if err != nil {
		switch err {
		case version.ErrNoVersionFound:
			return fmt.Errorf("no Gitlab Release tag found matching pattern %q of kind %q",
				g.versionFilter.Pattern,
				g.versionFilter.Kind,
			)
		default:
			return fmt.Errorf("filtering Gitlab release: %w", err)
		}
	}

	resultSource.Information = g.foundVersion.GetVersion()

	if len(resultSource.Information) == 0 {
		return fmt.Errorf("no Gitlab Release tag found matching pattern %q of kind %q",
			g.versionFilter.Pattern,
			g.versionFilter.Kind,
		)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("Gitlab release tag %q found matching pattern %q of kind %q",
		resultSource.Information,
		g.versionFilter.Pattern,
		g.versionFilter.Kind,
	)
	return nil

}
