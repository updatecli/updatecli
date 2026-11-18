package pypi

import (
	"slices"
	"strings"

	gopep440 "github.com/aquasecurity/go-pep440-version"
)

// release is a non-yanked release of a package.
type release struct {
	// version is the version as published, such as "1.0b2".
	version string
	// candidate is the version as compared by the versionfilter: normalized to semver, such
	// as "1.0.0-beta.2", unless the versionfilter kind is pep440.
	candidate string
	// status tells how the release relates to the age filter. Without age filter, every
	// release matches.
	status releaseAgeStatus
	// parsed is the PEP 440 version, parsed once for sorting, or nil when it doesn't parse.
	parsed *gopep440.Version
}

// newRelease returns a release, with its version parsed once for sorting.
func newRelease(ver, candidate string, status releaseAgeStatus) release {
	rel := release{version: ver, candidate: candidate, status: status}
	if parsed, err := gopep440.Parse(ver); err == nil {
		rel.parsed = &parsed
	}
	return rel
}

// releases holds the non-yanked releases of a package, by version as published.
// Versions are kept as published since normalization can map distinct releases,
// such as "1.0.0" and "1.0.0.post1", to the same candidate.
type releases map[string]release

// withStatus returns the releases having one of the given statuses, oldest first by PEP 440
// order. PyPI doesn't report the publication order, and some versionfilter kinds, such as
// "regex", take the last match as the newest, so a plain string sort would rank "1.9.0"
// above "1.10.0".
func (r releases) withStatus(statuses ...releaseAgeStatus) []release {
	var selected []release
	for _, rel := range r {
		if slices.Contains(statuses, rel.status) {
			selected = append(selected, rel)
		}
	}
	slices.SortFunc(selected, compareReleases)
	return selected
}

// count returns the number of releases having one of the given statuses.
func (r releases) count(statuses ...releaseAgeStatus) int {
	n := 0
	for _, rel := range r {
		if slices.Contains(statuses, rel.status) {
			n++
		}
	}
	return n
}

// versions returns the published versions of the releases having one of the given statuses,
// oldest first.
func (r releases) versions(statuses ...releaseAgeStatus) []string {
	var versions []string
	for _, rel := range r.withStatus(statuses...) {
		versions = append(versions, rel.version)
	}
	return versions
}

// candidates returns the candidates of the releases having one of the given statuses, oldest
// first and without duplicates, as distinct releases can share a candidate.
func (r releases) candidates(statuses ...releaseAgeStatus) []string {
	var candidates []string
	seen := map[string]bool{}
	for _, rel := range r.withStatus(statuses...) {
		if seen[rel.candidate] {
			continue
		}
		seen[rel.candidate] = true
		candidates = append(candidates, rel.candidate)
	}
	return candidates
}

// original returns the published version of a candidate matching the age filter. When several
// releases share the candidate, such as "1.0.0" and "1.0.0.post1" for "1.0.0", the highest by
// PEP 440 order is returned, so the newest build is used and the answer doesn't change from run
// to run.
func (r releases) original(candidate string) string {
	best := ""
	for _, rel := range r.withStatus(ageMatching) {
		if rel.candidate == candidate {
			best = rel.version
		}
	}

	if best == "" {
		return candidate
	}
	return best
}

// lookup returns the age status of a published version. A version given in its normalized
// form, such as "1.0.0-beta.2" for "1.0b2", is looked up through the releases normalizing to
// it, preferring one matching the age filter. The base release is preferred over post releases,
// whose normalization drops the ".postN" suffix: "1.0.0" designates "1.0.0" when it exists, and
// only a post release such as "1.0.post1" otherwise.
func (r releases) lookup(ver string) (status releaseAgeStatus, published bool) {
	if rel, found := r[ver]; found {
		return rel.status, true
	}

	for _, post := range []bool{false, true} {
		for _, rel := range r {
			if rel.candidate != ver || isPostRelease(rel.version) != post {
				continue
			}
			// Statuses are ordered by preference, ageMatching first.
			if !published || rel.status < status {
				status, published = rel.status, true
			}
		}
		if published {
			return status, published
		}
	}
	return status, published
}

// compareReleases compares two releases by PEP 440 order. A version that doesn't parse is
// ordered before those that do, and two such versions are compared as strings, so the order
// stays deterministic.
func compareReleases(a, b release) int {
	switch {
	case a.parsed == nil && b.parsed == nil:
		return strings.Compare(a.version, b.version)
	case a.parsed == nil:
		return -1
	case b.parsed == nil:
		return 1
	}

	if c := a.parsed.Compare(*b.parsed); c != 0 {
		return c
	}
	return strings.Compare(a.version, b.version)
}

// isPostRelease reports whether a PEP 440 version is a post release, such as "1.0.post1".
func isPostRelease(v string) bool {
	m := pep440PreRelease.FindStringSubmatch(v)
	return m != nil && m[2] == "post"
}
