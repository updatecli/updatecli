package release

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func (g *Gitea) Source(workingDir string) (string, error) {
	versions, err := g.SearchReleases()

	if err != nil {
		logrus.Error(err)
		return "", err
	}

	if len(versions) == 0 {
		logrus.Infof("%s No Gitea Release found. As a fallback you may be looking for git tags", result.ATTENTION)
		return "", errors.New("no result found")
	}

	g.foundVersion, err = g.spec.VersionFilter.Search(versions)

	if err != nil {
		switch err {
		case version.ErrNoVersionFound:
			logrus.Infof("%s No Gitea Release tag found matching pattern %q", result.FAILURE, g.versionFilter.Pattern)
			return "", errors.New("no result found")
		default:
			return "", err
		}
	}

	value := g.foundVersion.GetVersion()

	logrus.Infof("Latest Release found: %v", g.foundVersion.GetVersion())

	if len(value) == 0 {
		logrus.Infof("%s No Gitea Release tag found matching pattern %q", result.FAILURE, g.versionFilter.Pattern)
		return "", errors.New("no result found")
	} else if len(value) > 0 {
		logrus.Infof("%s Gitea Release tag %q found matching pattern %q", result.SUCCESS, value, g.versionFilter.Pattern)
		return value, nil
	}

	return "", fmt.Errorf("something unexpected happened in Gitea source")

}
