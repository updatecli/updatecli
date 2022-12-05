package npm

import (
	"errors"
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Condition checks that an Npm package version exist
func (n Npm) Condition(source string) (bool, error) {
	versionToCheck := n.spec.Version
	if versionToCheck == "" {
		versionToCheck = source
	}
	if len(versionToCheck) == 0 {
		return false, errors.New("no version defined")
	}

	_, versions, err := n.getVersions()
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

// ConditionFromSCM checks if a key exists in a yaml file
func (n Npm) ConditionFromSCM(source string, scm scm.ScmHandler) (bool, error) {
	return false, fmt.Errorf("SCM configuration is not supported for npm condition, aborting")
}
