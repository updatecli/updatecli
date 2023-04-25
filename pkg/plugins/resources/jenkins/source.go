package jenkins

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source returns the latest Jenkins version based on release type
func (j *Jenkins) Source(workingDir string, resultSource *result.Source) error {
	latest, versions, err := j.getVersions()
	if err != nil {
		return fmt.Errorf("searching jenkins version: %w", err)
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
		return fmt.Errorf("unknown version %s found for the %s release", j.spec.Version, j.spec.Release)
	}

	resultSource.Information = j.foundVersion
	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("version %q found for the Jenkins %s release", j.foundVersion, j.spec.Release)

	return nil
}
