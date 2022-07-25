package jenkins

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks that a Jenkins version exists and that the version
// match a valid release type
func (j Jenkins) Condition(source string) (bool, error) {

	versionToCheck := j.spec.Version
	// Override source input if version specified by parameter
	if versionToCheck == "" {
		versionToCheck = source
	}

	releaseType, err := ReleaseType(versionToCheck)
	if err != nil {
		return false, err
	}

	if releaseType != j.spec.Release {
		return false, fmt.Errorf(
			"wrong Release Type '%s' detected : Jenkins version '%s' is a '%s' release",
			j.spec.Release, versionToCheck, releaseType)
	}

	if len(versionToCheck) > 0 {
		_, versions, err := j.getVersions()
		if err != nil {
			return false, err
		}

		for _, v := range versions {
			if v == versionToCheck {
				fmt.Printf("%s %s release version '%s' available\n", result.SUCCESS, releaseType, versionToCheck)
				return true, nil
			}
		}
	}

	fmt.Printf("%s Version '%v' doesn't exist\n", result.FAILURE, versionToCheck)
	return false, nil
}

// ConditionFromSCM checks if a key exists in a yaml file
func (j Jenkins) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return false, fmt.Errorf("SCM configuration is not supported for Jenkins condition, aborting")
}
