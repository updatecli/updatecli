package branch

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

func (g *Gitlab) Source(workingDir string) (string, error) {
	versions, err := g.SearchBranches()

	if err != nil {
		return "", err
	}

	if len(versions) == 0 {
		logrus.Infof("%s No Gitlab branches found", result.FAILURE)
		return "", errors.New("no result found")
	}

	g.foundVersion, err = g.spec.VersionFilter.Search(versions)

	if err != nil {
		switch err {
		case version.ErrNoVersionFound:
			logrus.Infof("%s No Gitlab branches found matching pattern %q", result.FAILURE, g.versionFilter.Pattern)
			return "", errors.New("no result found")
		default:
			return "", err
		}
	}

	value := g.foundVersion.GetVersion()

	if len(value) == 0 {
		logrus.Infof("%s No Gitlab branches found matching pattern %q", result.FAILURE, g.versionFilter.Pattern)
		return "", errors.New("no result found")
	} else if len(value) > 0 {
		logrus.Infof("%s Gitlab branches %q found matching pattern %q", result.SUCCESS, value, g.versionFilter.Pattern)
		return value, nil
	}

	return "", fmt.Errorf("something unexpected happened in Gitlab source")

}
