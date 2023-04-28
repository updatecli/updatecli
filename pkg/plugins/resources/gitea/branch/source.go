package branch

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func (g *Gitea) Source(workingDir string, resultSource *result.Source) error {
	versions, err := g.SearchBranches()

	if err != nil {
		return fmt.Errorf("searching gitea branches: %w", err)
	}

	if len(versions) == 0 {
		logrus.Infof("%s No Gitea branches found", result.FAILURE)
		return errors.New("no gitea branches found")
	}

	g.foundVersion, err = g.spec.VersionFilter.Search(versions)

	if err != nil {
		switch err {
		case version.ErrNoVersionFound:
			return fmt.Errorf("no Gitea branches found matching pattern %q", g.versionFilter.Pattern)
		default:
			return fmt.Errorf("filtering gitea branches: %w", err)
		}
	}

	value := g.foundVersion.GetVersion()

	if len(value) == 0 {
		return fmt.Errorf("no Gitea branches found matching pattern %q", g.versionFilter.Pattern)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = value
	resultSource.Description = fmt.Sprintf("Gitea branches %q found matching pattern %q", value, g.versionFilter.Pattern)

	return nil

}
