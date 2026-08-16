package npm

import (
	"slices"
	"time"

	sv "github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
)

// filterVersionsByAge removes the versions which were published outside of the age
// window defined by the release age spec.
// publishDates is the `time` object returned by the npm registry, which maps a version
// to its release date. Versions without a valid release date are discarded.
func filterVersionsByAge(versions []string, publishDates map[string]string, releaseAge age.Spec) []string {
	if releaseAge.IsZero() {
		return versions
	}

	sanitizedVersions := []string{}

	for _, v := range versions {
		rawDate, found := publishDates[v]
		if !found {
			logrus.Debugf("ignoring version %q because the registry doesn't report any release date for it\n", v)
			continue
		}

		releaseDate, err := time.Parse(time.RFC3339, rawDate)
		if err != nil {
			logrus.Debugf("ignoring version %q due to invalid release date %q: %q\n", v, rawDate, err)
			continue
		}

		if releaseAge.Minimum != "" && releaseAge.IsOlderThan(releaseDate, nil) {
			logrus.Debugf("ignoring version %q because its age is below %q (released on %s)\n", v, releaseAge.Minimum, releaseDate)
			continue
		}

		if releaseAge.Maximum != "" && releaseAge.IsNewerThan(releaseDate, nil) {
			logrus.Debugf("ignoring version %q because its age is above %q (released on %s)\n", v, releaseAge.Maximum, releaseDate)
			continue
		}

		sanitizedVersions = append(sanitizedVersions, v)
	}

	return sanitizedVersions
}

// latestVersionMatchingAge returns the version to use when the versionfilter is of kind
// "latest" and an age filter discarded some versions.
// The npm dist-tag "latest" is used as soon as it matches the age window, otherwise we
// fallback to the most recently published version that does, based on the order in which
// the registry returned the versions.
// Because the dist-tag "latest" conventionally never points to a prerelease, the fallback
// ignores prereleases as well, unless the package only publishes those.
func latestVersionMatchingAge(distTagLatest string, orderedVersions, matchingVersions []string) string {
	if len(matchingVersions) == 0 {
		return ""
	}

	if slices.Contains(matchingVersions, distTagLatest) {
		return distTagLatest
	}

	logrus.Debugf("dist-tag latest %q doesn't match the age filter, looking for the most recently published version that does\n", distTagLatest)

	if v := lastPublishedVersion(orderedVersions, matchingVersions, true); v != "" {
		return v
	}

	if v := lastPublishedVersion(orderedVersions, matchingVersions, false); v != "" {
		logrus.Debugf("only prereleases match the age filter for that package\n")
		return v
	}

	// The registry didn't report the publication order, we can't tell which version
	// is the most recent one so we don't return any.
	logrus.Debugf("no publication order reported by the registry, ignoring the %d version(s) matching the age filter\n", len(matchingVersions))

	return ""
}

// lastPublishedVersion returns the last version of orderedVersions which is also part of
// matchingVersions, optionally skipping the prereleases.
func lastPublishedVersion(orderedVersions, matchingVersions []string, ignorePrerelease bool) string {
	for i := len(orderedVersions) - 1; i >= 0; i-- {
		if !slices.Contains(matchingVersions, orderedVersions[i]) {
			continue
		}

		if ignorePrerelease && isPrerelease(orderedVersions[i]) {
			logrus.Debugf("ignoring prerelease %q\n", orderedVersions[i])
			continue
		}

		return orderedVersions[i]
	}

	return ""
}

// isPrerelease returns true if the provided version is a semantic version with a
// prerelease identifier such as "6.0.0-dev.20250812".
// Versions that are not valid semantic versions are never considered as prereleases.
func isPrerelease(v string) bool {
	parsedVersion, err := sv.NewVersion(v)
	if err != nil {
		return false
	}

	return parsedVersion.Prerelease() != ""
}
