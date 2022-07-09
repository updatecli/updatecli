package release

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (g *Gitea) Source(workingDir string) (string, error) {
	versions, err := g.SearchReleases()

	if err != nil {
		logrus.Error(err)
		return "", err
	}

	if len(versions) == 0 {
		logrus.Infof("%s No Gitea Release found. As a fallback you may be looking for git tags", result.ATTENTION)
		return "", fmt.Errorf("no release found, exiting")
	}

	g.foundVersion, err = g.Spec.VersionFilter.Search(versions)
	if err != nil {
		return "", err
	}
	value := g.foundVersion.ParsedVersion

	logrus.Infof("Latest Release found: %v", g.foundVersion.ParsedVersion)

	if len(value) == 0 {
		logrus.Infof("%s No Gitea Release tag found matching pattern %q", result.FAILURE, g.versionFilter.Pattern)
		return "", nil
	} else if len(value) > 0 {
		logrus.Infof("%s Gitea Release tag %q found matching pattern %q", result.SUCCESS, value, g.versionFilter.Pattern)
		return value, nil
	}

	return "", fmt.Errorf("something unexpected happened in Gitea source")

}
