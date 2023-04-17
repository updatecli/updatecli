package language

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest go module version
func (g *Language) Source(workingDir string) (string, error) {
	_, err := g.versions()
	if err != nil {
		return "", err
	}

	if g.Version.GetVersion() != "" {
		logrus.Infof("%s Golang Version %s found", result.SUCCESS, g.Version.GetVersion())
		return g.Version.GetVersion(), nil
	}

	logrus.Infof("%s No Golang version found matching pattern %q", result.FAILURE, g.Spec.VersionFilter.Pattern)

	return "", fmt.Errorf("%s No Golang version found matching pattern %q ", result.FAILURE, g.Spec.VersionFilter.Pattern)
}
