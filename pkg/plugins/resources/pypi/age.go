package pypi

import (
	"time"

	gopep440 "github.com/aquasecurity/go-pep440-version"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
)

// releaseDate returns the earliest upload time among the non yanked files of a release.
// The earliest one is used so that a wheel uploaded later doesn't restart the cooldown.
// It returns false when the registry doesn't report any valid upload time.
func releaseDate(files []releaseFile) (time.Time, bool) {
	var earliest time.Time
	found := false

	for _, f := range files {
		if f.Yanked || f.UploadTime == "" {
			continue
		}

		t, err := time.Parse(time.RFC3339, f.UploadTime)
		if err != nil {
			log.Debugf("ignoring invalid upload time %q: %s", f.UploadTime, err)
			continue
		}

		if !found || t.Before(earliest) {
			earliest = t
			found = true
		}
	}

	return earliest, found
}

// releaseAgeStatus tells how a release relates to the age filter.
// The values are ordered by preference, from a release usable now to one never usable.
type releaseAgeStatus int

const (
	// ageMatching is a release inside the age window.
	ageMatching releaseAgeStatus = iota
	// ageTooRecent is a release younger than the minimum age, still cooling down.
	ageTooRecent
	// ageTooOld is a release older than the maximum age, which will never match again.
	ageTooOld
	// ageUndated is a release for which the registry reports no valid upload time.
	ageUndated
)

// ageStatus returns how a release relates to the age window.
// Only a release too recent is cooling down: the age filter will accept it later, unlike
// a release too old or one without upload time.
func ageStatus(ver string, files []releaseFile, releaseAge age.Spec) releaseAgeStatus {
	date, found := releaseDate(files)
	if !found {
		log.Debugf("ignoring version %q because the registry doesn't report any upload time for it", ver)
		return ageUndated
	}

	now := time.Now()
	if releaseAge.IsOlderThan(date, &now) {
		log.Debugf("ignoring version %q because its age is below %q (released on %s)", ver, releaseAge.Minimum, date)
		return ageTooRecent
	}
	if releaseAge.IsNewerThan(date, &now) {
		log.Debugf("ignoring version %q because its age is above %q (released on %s)", ver, releaseAge.Maximum, date)
		return ageTooOld
	}

	return ageMatching
}

// latestVersionMatchingAge returns the version to use when the versionfilter is of kind
// "latest" and the latest release reported by the registry doesn't match the age filter.
// It returns the highest stable version below latest by PEP 440 ordering, or the highest
// pre release below latest when the package only publishes those. Versions above latest
// are left out, so the fallback never goes past what the registry reports as latest.
func latestVersionMatchingAge(versions []string, latest string) string {
	var latestStable, latestPre *gopep440.Version
	var latestStableStr, latestPreStr string

	upperBound, err := gopep440.Parse(latest)
	if err != nil {
		log.Debugf("ignoring the fallback for latest version %q: %s", latest, err)
		return ""
	}

	for _, v := range versions {
		parsed, err := gopep440.Parse(v)
		if err != nil {
			log.Debugf("ignoring version %q: %s", v, err)
			continue
		}

		if !parsed.LessThan(upperBound) {
			continue
		}

		if parsed.IsPreRelease() {
			if latestPre == nil || parsed.GreaterThan(*latestPre) {
				latestPre = &parsed
				latestPreStr = v
			}
			continue
		}

		if latestStable == nil || parsed.GreaterThan(*latestStable) {
			latestStable = &parsed
			latestStableStr = v
		}
	}

	if latestStable != nil {
		return latestStableStr
	}

	return latestPreStr
}
