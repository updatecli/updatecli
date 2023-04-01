package gomod

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest go module version
func (g GoMod) Source(workingDir string) (string, error) {
	version, err := g.version(g.filename)
	if err != nil {
		return "", err
	}

	if version != "" {
		logrus.Infof("%s Version %s found for GO module %q", result.SUCCESS, version, g.spec.Module)
		return version, nil
	}

	logrus.Infof("%s No version found for module path %q", result.FAILURE, g.spec.Module)

	return "", fmt.Errorf("%s No version found for module path %q ", result.FAILURE, g.spec.Module)
}
