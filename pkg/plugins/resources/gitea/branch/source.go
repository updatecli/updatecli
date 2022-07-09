package branch

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

func (g *Gitea) Source(workingDir string) (string, error) {
	versions, err := g.SearchBranches()

	if err != nil {
		return "", err
	}

	if len(versions) == 0 {
		logrus.Infof("%s No Gitea branches found", result.ATTENTION)
		return "", nil
	}

	g.foundVersion, err = g.Spec.VersionFilter.Search(versions)
	if err != nil {
		return "", err
	}
	value := g.foundVersion.ParsedVersion

	if len(value) == 0 {
		logrus.Infof("%s No Gitea branches found matching pattern %q", result.FAILURE, g.versionFilter.Pattern)
		return "", nil
	} else if len(value) > 0 {
		logrus.Infof("%s Gitea branches %q found matching pattern %q", result.SUCCESS, value, g.versionFilter.Pattern)
		return value, nil
	}

	return "", fmt.Errorf("something unexpected happened in Gitea source")

}
