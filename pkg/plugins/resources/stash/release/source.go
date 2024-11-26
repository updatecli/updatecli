package release

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func (g *Stash) Source(workingDir string, resultSource *result.Source) error {
	versions, err := g.SearchReleases()

	if err != nil {
		logrus.Error(err)
		return fmt.Errorf("searching Bitbucket releases: %w", err)
	}

	if len(versions) == 0 {
		return fmt.Errorf("no Bitbucket Release found")
	}

	g.foundVersion, err = g.spec.VersionFilter.Search(versions)

	if err != nil {
		switch err {
		case version.ErrNoVersionFound:
			return fmt.Errorf("no Bitbucket release tag found matching pattern %q", g.versionFilter.Pattern)
		default:
			return fmt.Errorf("filtering Bitbucket release: %w", err)
		}
	}

	value := g.foundVersion.GetVersion()

	if len(value) == 0 {
		return fmt.Errorf("no Bitbucket release tag found matching pattern %q of kind %q", g.versionFilter.Pattern, g.versionFilter.Kind)
	}

	resultSource.Information = []result.SourceInformation{{
		Key:   resultSource.ID,
		Value: value,
	}}
	resultSource.Description = fmt.Sprintf("Bitbucket release tag %q found matching pattern %q", value, g.versionFilter.Pattern)
	resultSource.Result = result.SUCCESS

	return nil

}
