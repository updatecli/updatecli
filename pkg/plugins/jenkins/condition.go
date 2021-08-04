package jenkins

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/scm"
)

// Condition checks that a Jenkins version exists and that the version
// match a valid release type
func (j *Jenkins) Condition(source string) (bool, error) {

	versionExist := false
	validReleaseType := false

	// Override source input if version specified by parameter
	if len(j.Version) == 0 {
		j.Version = source
	}

	releaseType, err := ReleaseType(j.Version)
	if err != nil {
		return false, err
	}

	err = j.Validate()

	if err != nil {
		return false, err
	}

	validReleaseType = true

	if strings.Compare(releaseType, j.Release) != 0 {
		return false, fmt.Errorf(
			"Wrong Release Type '%s' detected : Jenkins version '%s' is a '%s' release",
			j.Release, j.Version, releaseType)
	}

	if len(j.Version) > 0 {
		_, versions, err := GetVersions()

		if err != nil {
			return false, err
		}
		for _, v := range versions {
			if strings.Compare(v, j.Version) == 0 {
				versionExist = true
				break
			}
		}
	}
	if !versionExist {
		fmt.Printf("\u2717 Version '%v' doesn't exist\n", j.Version)
		return false, nil

	} else if !validReleaseType {
		fmt.Printf("\u2717 Wrong Release Type: %v for version %v\n", releaseType, j.Version)
		return false, nil
	}

	fmt.Printf("\u2714 %s release version '%s' available\n", releaseType, j.Version)

	return true, nil
}

// ConditionFromSCM checks if a key exists in a yaml file
func (j *Jenkins) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	return false, fmt.Errorf("SCM configuration is not supported for Jenkins condition, aborting")
}
