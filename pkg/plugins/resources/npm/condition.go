package npm

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition checks that an Npm package version exist
func (n Npm) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	if scm != nil {
		logrus.Warningf("SCM configuration is not supported for npm condition, aborting")

	}

	versionToCheck := n.spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}
	if len(versionToCheck) == 0 {
		return false, "", errors.New("no version defined")
	}

	_, versions, err := n.getVersions()
	if err != nil {
		return false, "", err
	}

	for _, v := range versions {
		if v == versionToCheck {
			return true, fmt.Sprintf("release version %q available", versionToCheck), nil
		}
	}

	return false, fmt.Sprintf("Version %q doesn't exist\n", versionToCheck), nil
}
