package branch

import (
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func (g *Stash) Source(workingDir string, resultSource *result.Source) error {
	versions, err := g.SearchBranches()

	if err != nil {
		return fmt.Errorf("searching Bitbucket branches: %w", err)
	}

	if len(versions) == 0 {
		return errors.New("no Bitbucket branch found")
	}

	g.foundVersion, err = g.spec.VersionFilter.Search(versions)

	if err != nil {
		switch err {
		case version.ErrNoVersionFound:
			return fmt.Errorf("no Bitbucket branches found matching pattern %q", g.versionFilter.Pattern)
		default:
			return fmt.Errorf("filtering branches: %w", err)
		}
	}

	value := g.foundVersion.GetVersion()

	if len(value) == 0 {
		return fmt.Errorf("no Bitbucket branches found matching pattern %q", g.versionFilter.Pattern)
	}

	resultSource.Information = value
	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("Bitbucket branches %q found matching pattern %q", value, g.versionFilter.Pattern)

	return nil
}
