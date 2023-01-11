package cargopackage

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// ConditionFromSCM returns an error because it's not supported
func (cp *CargoPackage) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	logrus.Warningf("SCM configuration is not supported for condition of type cargopackage. Remove the `scm` directive from condition to remove this warning message")
	return cp.Condition(source)
}

// Condition checks if a cargo package with a specific version is published
// We assume that if we can't find the package version in the index, then it means it doesn't exist.
func (cp *CargoPackage) Condition(source string) (bool, error) {
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
			fmt.Printf("%s release version '%s' available\n", result.SUCCESS, versionToCheck)
			return true, nil
		}
	}

	fmt.Printf("%s Version %q doesn't exist\n", result.FAILURE, versionToCheck)

	return false, nil
}
