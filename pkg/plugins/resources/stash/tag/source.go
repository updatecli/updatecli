package tag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func (g *Stash) Source(workingDir string, resultSource *result.Source) error {
	versions, err := g.SearchTags()

	if err != nil {
		logrus.Error(err)
		return fmt.Errorf("searching Bitbucket tags: %w", err)
	}

	if len(versions) == 0 {
		return fmt.Errorf("no Bitbucket tags found")
	}

	g.foundVersion, err = g.spec.VersionFilter.Search(versions)

	if err != nil {
		switch err {
		case version.ErrNoVersionFound:
			return fmt.Errorf("no Bitbucket tags found matching pattern %q", g.versionFilter.Pattern)
		default:
			return fmt.Errorf("filtering Bitbucket tags: %w", err)
		}
	}

	value := g.foundVersion.GetVersion()

	if len(value) == 0 {
		return fmt.Errorf("no Bitbucket tags found matching pattern %q", g.versionFilter.Pattern)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = []result.SourceInformation{{
		Key:   resultSource.ID,
		Value: value,
	}}
	resultSource.Description = fmt.Sprintf("Bitbucket tag %q found matching pattern %q", value, g.versionFilter.Pattern)

	return nil

}
