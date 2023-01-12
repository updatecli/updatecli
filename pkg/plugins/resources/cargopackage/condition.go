package cargopackage

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks that a git tag exists
func (cp *CargoPackage) Condition(source string) (bool, error) {
	return cp.condition(source)
}

// ConditionFromSCM test if a tag exists from a git repository specific from SCM
func (cp *CargoPackage) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	path := scm.GetDirectory()

	if cp.spec.IndexDir != "" {
		logrus.Warningf("IndexDir is defined and set to %q but is overridden by the scm definition %q",
			cp.spec.IndexDir,
			path)
	}
	if cp.spec.IndexUrl != "" {
		logrus.Warningf("IndexUrl is defined and set to %q but is overridden by the scm definition %q",
			cp.spec.IndexDir,
			path)
	}
	cp.indexDir = path

	return cp.condition(source)
}

// Condition checks if a cargo package with a specific version is published
// We assume that if we can't find the package version in the index, then it means it doesn't exist.
func (cp *CargoPackage) condition(source string) (bool, error) {
	versionToCheck := cp.spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}
	if len(versionToCheck) == 0 {
		return false, errors.New("no version defined")
	}

	_, versions, err := cp.getVersions()
	if err != nil {
		return false, err
	}

	for _, v := range versions {
		if v == versionToCheck {
			logrus.Infof("%s release version '%s' available\n", result.SUCCESS, versionToCheck)
			return true, nil
		}
	}

	logrus.Infof("%s Version %q doesn't exist\n", result.FAILURE, versionToCheck)

	return false, nil
}
