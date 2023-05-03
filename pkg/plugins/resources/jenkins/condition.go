package jenkins

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks that a Jenkins version exists and that the version
// match a valid release type
func (j Jenkins) Condition(source string, scm scm.ScmHandler, resultCondition *result.Condition) error {

	if scm != nil {
		logrus.Warningf("SCM configuration is not supported for Jenkins condition")
	}

	versionToCheck := j.spec.Version
	// Override source input if version specified by parameter
	if versionToCheck == "" {
		versionToCheck = source
	}

	releaseType, err := ReleaseType(versionToCheck)
	if err != nil {
		return err
	}

	if releaseType != j.spec.Release {
		return fmt.Errorf(
			"wrong Release Type '%s' detected : Jenkins version '%s' is a '%s' release",
			j.spec.Release, versionToCheck, releaseType)
	}

	if len(versionToCheck) > 0 {
		_, versions, err := j.getVersions()
		if err != nil {
			return err
		}

		for _, v := range versions {
			if v == versionToCheck {
				resultCondition.Result = result.SUCCESS
				resultCondition.Pass = true
				resultCondition.Description = fmt.Sprintf("%s release version %q available\n", releaseType, versionToCheck)
				return nil
			}
		}
	}

	resultCondition.Result = result.FAILURE
	resultCondition.Pass = false
	resultCondition.Description = fmt.Sprintf("version %q doesn't exist", versionToCheck)
	return nil
}
