package gomodule

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest go module version
func (g GoModule) Source(workingDir string) (string, error) {
	version, _, err := g.versions()
	if err != nil {
		return "", err
	}

	if version != "" {
		logrus.Infof("%s Version %s found for the GO module %q", result.SUCCESS, version, g.spec.Path)
		return version, nil
	}

	logrus.Infof("%s Unknown version %s found for GO module %q ", result.FAILURE, version, g.spec.Path)

	return "", fmt.Errorf("%s Unknown version %s found for GO module %q ", result.FAILURE, version, g.spec.Path)
}
