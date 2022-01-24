package jenkins

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source return the latest Jenkins version based on release type
func (j Jenkins) Source(workingDir string) (string, error) {
	latest, versions, err := j.getVersions()
	if err != nil {
		return "", err
	}

	switch j.spec.Release {
	case WEEKLY:
		fmt.Printf("%s Version %s found for the %s release", result.SUCCESS, latest, WEEKLY)
		return latest, nil
	case STABLE:
		vs := []*semver.Version{}
		for _, r := range versions {
			v, err := semver.StrictNewVersion(r)
			if err != nil {
				// This version can be ignored. Let's jump to the next one
				continue
			}
			vs = append(vs, v)
		}

		sort.Sort(semver.Collection(vs))
		found := vs[len(vs)-1]
		versionFound := found.Original()

		fmt.Printf("%s Version %s found for the Jenkins %s release", result.SUCCESS, versionFound, j.spec.Release)
		return versionFound, nil
	default:
		fmt.Printf("%s Unknown version %s found for the %s release", result.FAILURE, j.spec.Version, j.spec.Release)
		return "unknown", fmt.Errorf("Unknown Jenkins version found")
	}
}
