package jenkins

import (
	"fmt"
	"strings"
)

// Source return the latest Jenkins version based on release type
func (j *Jenkins) Source(workingDir string) (string, error) {
	err := j.Validate()

	if err != nil {
		return "", err
	}

	latest, versions, err := GetVersions()

	if err != nil {
		return "", err
	}

	if strings.Compare(WEEKLY, j.Release) == 0 {
		fmt.Printf("\u2714 Version %s found for the %s release", latest, WEEKLY)
		return latest, nil
	}

	if strings.Compare(STABLE, j.Release) == 0 {
		found := filter(versions, func(s string) bool {
			v := NewVersion(s)
			return v.Patch != ""
		})
		fmt.Printf("\u2714 Version %s found for the Jenkins %s release", found[len(found)-1], j.Release)
		return found[len(found)-1], nil

	}

	fmt.Printf("\u2717 Unknown version %s found for the %s release", j.Version, j.Release)

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
