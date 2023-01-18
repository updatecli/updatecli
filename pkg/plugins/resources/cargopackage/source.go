package cargopackage

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest npm package version
func (cp CargoPackage) Source(workingDir string) (string, error) {
	logrus.Debugf("Registy RootDir: %s, workingDir: %s", cp.registry.RootDir, workingDir)
	if cp.isSCM {
		// We are in a scm context, workingDir is holding the data
		cp.registry.RootDir = workingDir
	}

	version, _, err := cp.getVersions()
	if err != nil {
		return "", err
	}

	if version != "" {
		logrus.Infof("%s Version %s found for package name %q", result.SUCCESS, version, cp.spec.Package)
		return version, nil
	}

	logrus.Infof("%s Unknown version %s found for package name %s ", result.FAILURE, version, cp.spec.Package)

	return "", fmt.Errorf("%s Unknown version %s found for package name %s ", result.FAILURE, version, cp.spec.Package)
}
