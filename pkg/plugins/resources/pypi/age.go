package pypi

import (
	"errors"
	"fmt"
	"time"

	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

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

// releaseDate returns the earliest upload time among the non yanked files of a release.
// The earliest one is used so that a wheel uploaded later doesn't restart the cooldown.
// It returns false when the registry doesn't report any valid upload time.
func releaseDate(files []releaseFile) (time.Time, bool) {
	var earliest time.Time
	found := false

	for _, f := range files {
		if f.Yanked {
			continue
		}

		t, ok := uploadTime(f)
		if !ok {
			continue
		}

		if !found || t.Before(earliest) {
			earliest = t
			found = true
		}
	}

	return earliest, found
}

// uploadTime returns the upload time of a file, from "upload_time_iso_8601" or, for a registry
// only reporting the older "upload_time", from that field read as UTC.
func uploadTime(f releaseFile) (time.Time, bool) {
	if f.UploadTime != "" {
		t, err := time.Parse(time.RFC3339, f.UploadTime)
		if err == nil {
			return t, true
		}
		log.Debugf("ignoring invalid upload time %q: %s", f.UploadTime, err)
	}

	if f.LegacyUploadTime != "" {
		t, err := time.Parse(legacyUploadTimeLayout, f.LegacyUploadTime)
		if err == nil {
			return t, true
		}
		t, errRFC3339 := time.Parse(time.RFC3339, f.LegacyUploadTime)
		if errRFC3339 == nil {
			return t, true
		}
		log.Debugf("ignoring invalid upload time %q: %s", f.LegacyUploadTime, errors.Join(err, errRFC3339))
	}

	return time.Time{}, false
}

// ageStatus returns how a release relates to the age window.
// Only a release too recent is cooling down: the age filter will accept it later, unlike
// a release too old or one without upload time. now is the instant every release is judged
// against, taken once so that releases close to a boundary are judged consistently.
func ageStatus(ver string, files []releaseFile, releaseAge age.Spec, now time.Time) releaseAgeStatus {
	date, found := releaseDate(files)
	if !found {
		log.Debugf("ignoring version %q because the registry doesn't report any upload time for it", ver)
		return ageUndated
	}

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
//
// It keeps the releases ranking below latest by PEP 440 order, such as "2.0rc1" for "2.0" or
// "1.0" for "1.0+abc", but not "1.0+zzz" for "1.0+abc", then searches them as the "pep440" kind
// does: the highest stable release or, when none is stable, the highest pre release. Comparing
// the parsed releases rather than building a specifier from latest works for any version the
// registry reports as latest, so the fallback never goes past it.
//
// It returns an empty string when no release qualifies, or when latest isn't a PEP 440 version,
// such as "2004d1", as nothing can then be ranked below it. It returns an error when the search
// fails for another reason, so that a failure isn't reported as an age hold.
func latestVersionMatchingAge(matching []release, latest release) (string, error) {
	if latest.parsed == nil {
		log.Debugf("no fallback for the latest version %q, which isn't a PEP 440 version", latest.version)
		return "", nil
	}

	var below []string
	for _, rel := range matching {
		if rel.parsed != nil && rel.parsed.LessThan(*latest.parsed) {
			below = append(below, rel.version)
		}
	}

	if len(below) == 0 {
		return "", nil
	}

	filter := version.Filter{Kind: version.PEP440VERSIONKIND, Pattern: "*"}
	found, err := filter.Search(below)
	switch {
	case err == nil:
		return found.GetVersion(), nil
	case errors.Is(err, version.ErrNoVersionFound),
		errors.Is(err, version.ErrNoVersionsFound),
		errors.Is(err, version.ErrNoValidPep440VersionFound):
		log.Debugf("no release below the latest version %q matches the age filter: %s", latest.version, err)
		return "", nil
	default:
		return "", fmt.Errorf("looking for a release below the latest version %q: %w", latest.version, err)
	}
}
