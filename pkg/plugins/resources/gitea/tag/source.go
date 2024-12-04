package tag

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func (g *Gitea) Source(workingDir string, resultSource *result.Source) error {
	versions, err := g.SearchTags()

	if err != nil {
		logrus.Error(err)
		return fmt.Errorf("search gitea tags: %w", err)
	}

	if len(versions) == 0 {
		return fmt.Errorf("no Gitea Tags found")
	}

	g.foundVersion, err = g.spec.VersionFilter.Search(versions)

	if err != nil {
		switch err {
		case version.ErrNoVersionFound:
			return fmt.Errorf("no Gitea tags found matching pattern %q", g.versionFilter.Pattern)
		default:
			return fmt.Errorf("no Gitea tags found matching pattern %q: %w", g.versionFilter.Pattern, err)
		}
	}

	value := g.foundVersion.GetVersion()

	if len(value) == 0 {
		return fmt.Errorf("no Gitea tags found matching pattern %q", g.versionFilter.Pattern)
	}

	resultSource.Result = result.SUCCESS
	resultSource.Information = []result.SourceInformation{{
		Key:   resultSource.ID,
		Value: value,
	}}
	resultSource.Description = fmt.Sprintf("Gitea tags %q found matching pattern %q of kind %q",
		value,
		g.versionFilter.Pattern,
		g.versionFilter.Kind,
	)

	return nil

}
