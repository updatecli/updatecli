package branch

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func (g *Gitlab) Source(workingDir string, resultSource *result.Source) error {
	versions, err := g.SearchBranches()

	if err != nil {
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
