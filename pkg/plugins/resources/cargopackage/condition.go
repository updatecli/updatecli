package cargopackage

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition checks if a cargo package with a specific version is published
// We assume that if we can't find the package version in the index, then it means it doesn't exist.
func (cp *CargoPackage) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	if scm != nil {
		path := scm.GetDirectory()
		if cp.spec.Registry.RootDir != "" {
			logrus.Warningf("Registry.RootDir is defined and set to %q but is overridden by the scm definition %q",
				cp.spec.Registry.RootDir,
				path)
		}
		if cp.spec.Registry.URL != "" {
			logrus.Warningf("Registry.URL is defined and set to %q but is overridden by the scm definition %q",
				cp.spec.Registry.URL,
				path)
		}
		cp.registry.RootDir = path
	}

	versionToCheck := cp.spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}
	if len(versionToCheck) == 0 {
		return false, "", errors.New("no version defined")
	}

	_, versions, err := cp.getVersions()
	if err != nil {
		return false, "", fmt.Errorf("getting cargo package version: %w", err)
	}

	for _, v := range versions {
		if v == versionToCheck {
			return true, fmt.Sprintf("release version %q available\n", versionToCheck), nil
		}
	}

	return false, fmt.Sprintf("version %q doesn't exist\n", versionToCheck), nil
}
