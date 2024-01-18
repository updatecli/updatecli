package tag

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func (g *Gitlab) Source(workingDir string, resultSource *result.Source) error {
	versions, err := g.SearchTags()

	if err != nil {
		return fmt.Errorf("searching GitLab tags: %w", err)
	}

	if len(versions) == 0 {
		return fmt.Errorf("no GitLab Tags found")
	}

	g.foundVersion, err = g.spec.VersionFilter.Search(versions)

	if err != nil {
		switch err {
		case version.ErrNoVersionFound:
			return fmt.Errorf("no GitLab tags found matching pattern %q of kind %q",
				g.versionFilter.Pattern,
				g.versionFilter.Kind,
			)
		default:
			return fmt.Errorf("filtering tag: %w", err)
		}
	}

	resultSource.Information = g.foundVersion.GetVersion()

	if len(resultSource.Information) == 0 {
		return fmt.Errorf("no GitLab tags found matching pattern %q of kind %q",
			g.versionFilter.Pattern,
			g.versionFilter.Kind,
		)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("GitLab tags %q found matching pattern %q of kind %q",
		resultSource.Information,
		g.versionFilter.Pattern,
		g.versionFilter.Kind,
	)

	return nil

}
