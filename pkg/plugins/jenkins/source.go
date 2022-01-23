package jenkins

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source return the latest Jenkins version based on release type
func (j Jenkins) Source(workingDir string) (string, error) {
	latest, versions, err := GetVersions()
	if err != nil {
		return "", err
	}

	if strings.Compare(WEEKLY, j.spec.Release) == 0 {
		fmt.Printf("%s Version %s found for the %s release", result.SUCCESS, latest, WEEKLY)
		return latest, nil
	}

	if strings.Compare(STABLE, j.spec.Release) == 0 {
		found := filter(versions, func(s string) bool {
			v := NewVersion(s)
			return v.Patch != ""
		})
		fmt.Printf("%s Version %s found for the Jenkins %s release", result.SUCCESS, found[len(found)-1], j.spec.Release)
		return found[len(found)-1], nil

	}

	fmt.Printf("%s Unknown version %s found for the %s release", result.FAILURE, j.spec.Version, j.spec.Release)

	return "unknown", fmt.Errorf("Unknown Jenkins version found")
}

func filter(ss []string, test func(string) bool) (ret []string) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}
