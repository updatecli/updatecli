package npm

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks that an Npm package version exist
func (n Npm) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {

	if scm != nil {
		logrus.Warningf("SCM configuration is not supported for npm condition, aborting")

	}

	versionToCheck := n.spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}
	if len(versionToCheck) == 0 {
		return errors.New("no version defined")
	}

	_, versions, err := n.getVersions()
	if err != nil {
		return err
	}

	for _, v := range versions {
		if v == versionToCheck {
			resultCondition.Pass = true
			resultCondition.Result = result.SUCCESS
			resultCondition.Description = fmt.Sprintf("release version %q available", versionToCheck)
			return nil
		}
	}

	resultCondition.Result = result.FAILURE
	resultCondition.Pass = false
	resultCondition.Description = fmt.Sprintf("Version %q doesn't exist\n", versionToCheck)

	return nil
}
