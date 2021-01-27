package jenkins

import (
	"fmt"
	"strings"

	"github.com/olblak/updateCli/pkg/core/scm"
)

// Condition checks that a Jenkins version exists and that the version
// match a valid release type
func (j *Jenkins) Condition(source string) (bool, error) {

	version := source
	versionExist := false
	validReleaseType := false

	// Override source input if version specified by parameter
	if len(j.Version) > 0 {
		version = j.Version
	}

	releaseType, err := ReleaseType(version)
	if err != nil {
		return false, err
	}

	validReleaseType = true

	if len(j.Release) > 0 {
		if strings.Compare(releaseType, j.Release) != 0 {
			return false, fmt.Errorf(
				"Wrong Release Type '%s' detected : Jenkins version '%s' is a '%s' release",
				j.Release, version, releaseType)
		}
	}

	if len(version) > 0 {
		_, versions, err := GetVersions()

		if err != nil {
			return false, err
		}
		for _, v := range versions {
			if strings.Compare(v, version) == 0 {
				versionExist = true
				break
			}
		}
	}
	if !versionExist {
		fmt.Printf("\u2717 Version '%v' doesn't exist\n", version)
		return false, nil

	} else if !validReleaseType {
		fmt.Printf("\u2717 Wrong Release Type: %v for version %v\n", releaseType, version)
		return false, nil
	}

	fmt.Printf("\u2714 %s release version '%s' available\n", releaseType, version)

	return true, nil
}

// ConditionFromSCM checks if a key exists in a yaml file
func (j *Jenkins) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	return false, fmt.Errorf("SCM configuration is not supported for Jenkins condition, aborting")
}
