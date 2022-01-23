package jenkins

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks that a Jenkins version exists and that the version
// match a valid release type
func (j Jenkins) Condition(source string) (bool, error) {

	versionExist := false
	validReleaseType := false

	// Override source input if version specified by parameter
	if len(j.spec.Version) == 0 {
		j.spec.Version = source
	}

	releaseType, err := ReleaseType(j.spec.Version)
	if err != nil {
		return false, err
	}

	validReleaseType = true

	if strings.Compare(releaseType, j.spec.Release) != 0 {
		return false, fmt.Errorf(
			"Wrong Release Type '%s' detected : Jenkins version '%s' is a '%s' release",
			j.spec.Release, j.spec.Version, releaseType)
	}

	if len(j.spec.Version) > 0 {
		_, versions, err := GetVersions()

		if err != nil {
			return false, err
		}
		for _, v := range versions {
			if strings.Compare(v, j.spec.Version) == 0 {
				versionExist = true
				break
			}
		}
	}
	if !versionExist {
		fmt.Printf("%s Version '%v' doesn't exist\n", result.FAILURE, j.spec.Version)
		return false, nil

	} else if !validReleaseType {
		fmt.Printf("%s Wrong Release Type: %v for version %v\n", result.FAILURE, releaseType, j.spec.Version)
		return false, nil
	}

	fmt.Printf("%s %s release version '%s' available\n", result.SUCCESS, releaseType, j.spec.Version)

	return true, nil
}

// ConditionFromSCM checks if a key exists in a yaml file
func (j Jenkins) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return false, fmt.Errorf("SCM configuration is not supported for Jenkins condition, aborting")
}
