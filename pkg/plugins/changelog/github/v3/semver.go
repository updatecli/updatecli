package changelog

import (
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v69/github"
)

// sortReleases tries to sort releases by their tag name.
func sortReleasesBySemver(releases *[]*github.RepositoryRelease) {

	var semverReleases []*github.RepositoryRelease

	for _, release := range *releases {
		_, err := semver.NewVersion(release.GetTagName())
		if err == nil {
			semverReleases = append(semverReleases, release)
		}
	}

	sort.SliceStable(semverReleases, func(i, j int) bool {

		versionI, err := semver.NewVersion(semverReleases[i].GetTagName())
		if err != nil {
			return false
		}

		versionJ, err := semver.NewVersion(semverReleases[j].GetTagName())
		if err != nil {
			return false
		}

		return versionJ.LessThan(versionI)
	})

	*releases = semverReleases
}

// isSemverDetected tries to detect if our version range are semver compliant
func isSemverDetected(from, to string) bool {

	_, err := semver.NewVersion(from)
	if err != nil {
		return false
	}

	_, err = semver.NewVersion(to)

	return err == nil
}
