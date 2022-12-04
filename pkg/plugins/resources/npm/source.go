package npm

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest npm package version
func (n Npm) Source(workingDir string) (string, error) {
	version, _, err := n.getVersions()
	if err != nil {
		return "", err
	}

	if version != "" {
		logrus.Infof("%s Version %s found for package name %q", result.SUCCESS, version, n.spec.Name)
		return version, nil
	}

	logrus.Infof("%s Unknown version %s found for package name %s ", result.FAILURE, version, n.spec.Name)

	return "", fmt.Errorf("%s Unknown version %s found for package name %s ", result.FAILURE, version, n.spec.Name)
}
