package jenkins

import (
	"fmt"
	"strings"
)

// Source return the latest Jenkins version based on release type
func (j *Jenkins) Source(workingDir string) (string, error) {
	latest, versions, err := GetVersions()

	if err != nil {
		return "", err
	}

	if len(j.Release) == 0 || strings.Compare(WEEKLY, j.Release) == 0 {
		fmt.Printf("\u2714 Version %s found for the %s release", latest, WEEKLY)
		return latest, nil
	}

	if strings.Compare(STABLE, j.Release) == 0 {
		found := filter(versions, func(s string) bool {
			v := NewVersion(s)
			return v.Patch != ""
		})
		fmt.Println(found[len(found)-1])
		return found[len(found)-1], nil

	}

	if len(j.Release) == 0 {

		splitIdentifier := strings.Split(j.Version, ".")
		if len(splitIdentifier) > -1 {
			id := NewVersion(j.Version)
			// In this case we assume that we provided a valid version
			found := filter(versions, func(s string) bool {
				v := NewVersion(s)

				switch len(splitIdentifier) {
				case 0:
					return id.Major == v.Major
				case 1:
					return id.Major == v.Major && id.Minor == v.Minor
				case 2:
					return id.Major == v.Major && id.Minor == v.Minor && id.Patch == v.Patch
				default:
					return false
				}
			})
			fmt.Printf("%s requested, filtered list to %s", j.Version, found)
		}
	}

	j.Version = latest

	fmt.Printf("\u2714 Version %s found for the %s release", j.Version, j.Release)

	return latest, nil
}

func filter(ss []string, test func(string) bool) (ret []string) {
	for _, s := range ss {
		if test(s) {
			ret = append(ret, s)
		}
	}
	return
}
