package jenkins

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest Jenkins version based on release type
func (j *Jenkins) Source(workingDir string) (string, error) {
	latest, versions, err := j.getVersions()
	if err != nil {
		return "", err
	}

	switch j.spec.Release {
	case WEEKLY:
		j.foundVersion = latest
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
		j.foundVersion = found.Original()
	default:
		fmt.Printf("%s Unknown version %s found for the %s release", result.FAILURE, j.spec.Version, j.spec.Release)
		return "unknown", fmt.Errorf("unknown Jenkins version found")
	}

	fmt.Printf("%s Version %s found for the Jenkins %s release", result.SUCCESS, j.foundVersion, j.spec.Release)
	return j.foundVersion, nil
}
