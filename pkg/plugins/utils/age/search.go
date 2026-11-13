package age

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Search returns the version matching the filter among the versions kept by the age filter.
//
// dated holds the published versions whose release date is known, so it contains kept. When
// none of the kept versions matches but one of dated does, every version the filter accepts is
// outside the age window. It returns ErrNoVersionMatchingAge, so that callers skip rather than
// fail, as a version filter starting from the current version, such as ">=1.5.0", finds nothing
// while that version is too recent. A version without a known release date isn't counted, as
// the age filter can't be what discarded it. Any other error is returned as is.
func Search(filter *version.Filter, dated, kept []string) (version.Version, error) {
	found, err := filter.Search(kept)
	if err == nil {
		return found, nil
	}

	if !HeldBack(filter, missingFrom(dated, kept)) {
		return found, err
	}

	return found, fmt.Errorf("%w: every version matching %q is outside the age window", ErrNoVersionMatchingAge, filter.Pattern)
}

// HeldBack returns true if the filter matches one of the versions the age filter discarded.
// When the filter matches none of the versions it kept, the age filter is then what prevents
// a match, an age hold rather than a missing version.
func HeldBack(filter *version.Filter, discarded []string) bool {
	return len(discarded) > 0 && filter.Matches(discarded)
}

// MatchesUndated returns true if the filter matches one of the published versions whose release
// date is unknown, that is missing from dated. The age filter can't be applied to such a version,
// so when no dated version matches, the source is skipped rather than failed, as it can't tell
// whether a version would match once dated.
func MatchesUndated(filter *version.Filter, published, dated []string) bool {
	undated := missingFrom(published, dated)
	return len(undated) > 0 && filter.Matches(undated)
}

// missingFrom returns the versions of all that aren't in some.
func missingFrom(all, some []string) []string {
	isIn := make(map[string]bool, len(some))
	for _, v := range some {
		isIn[v] = true
	}

	var missing []string
	for _, v := range all {
		if !isIn[v] {
			missing = append(missing, v)
		}
	}

	return missing
}
