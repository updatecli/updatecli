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
}

// releases holds the non-yanked releases of a package, by version as published.
// Versions are kept as published since normalization can map distinct releases,
// such as "1.0.0" and "1.0.0.post1", to the same candidate.
type releases map[string]release

// withStatus returns the releases having one of the given statuses, sorted by version.
func (r releases) withStatus(statuses ...releaseAgeStatus) []release {
	var selected []release
	for _, rel := range r {
		if slices.Contains(statuses, rel.status) {
			selected = append(selected, rel)
		}
	}
	slices.SortFunc(selected, func(a, b release) int { return strings.Compare(a.version, b.version) })
	return selected
}

// versions returns the published versions of the releases having one of the given statuses, sorted.
func (r releases) versions(statuses ...releaseAgeStatus) []string {
	var versions []string
	for _, rel := range r.withStatus(statuses...) {
		versions = append(versions, rel.version)
	}
	return versions
}

// candidates returns the candidates of the releases having one of the given statuses, sorted
// and without duplicates, as distinct releases can share a candidate.
func (r releases) candidates(statuses ...releaseAgeStatus) []string {
	var candidates []string
	for _, rel := range r.withStatus(statuses...) {
		candidates = append(candidates, rel.candidate)
	}
	slices.Sort(candidates)
	return slices.Compact(candidates)
}

// original returns the published version of a candidate matching the age filter. When several
// releases share the candidate, such as "1.0.0" and "1.0.0.post1" for "1.0.0", the base release
// is preferred, then the highest post release, so the answer doesn't change from run to run.
func (r releases) original(candidate string) string {
	var best string
	var bestParsed gopep440.Version

	for _, rel := range r.withStatus(ageMatching) {
		if rel.candidate != candidate {
			continue
		}
		if !isPostRelease(rel.version) {
			return rel.version
		}

		parsed, err := gopep440.Parse(rel.version)
		if err != nil {
			continue
		}
		if best == "" || parsed.GreaterThan(bestParsed) {
			best, bestParsed = rel.version, parsed
		}
	}

	if best == "" {
		return candidate
	}
	return best
}

// lookup returns the age status of a published version. A version given in its normalized
// form, such as "1.0.0-beta.2" for "1.0b2", is looked up through the releases normalizing to
// it, preferring one matching the age filter. Post releases are left out of that lookup:
// their normalization drops the ".postN" suffix, so "1.0.0" designates the base release,
// not "1.0.0.post1".
func (r releases) lookup(ver string) (status releaseAgeStatus, published bool) {
	if rel, found := r[ver]; found {
		return rel.status, true
	}

	for _, rel := range r {
		if rel.candidate != ver || isPostRelease(rel.version) {
			continue
		}
		// Statuses are ordered by preference, ageMatching first.
		if !published || rel.status < status {
			status, published = rel.status, true
		}
	}
	return status, published
}

// isPostRelease reports whether a PEP 440 version is a post release, such as "1.0.post1".
func isPostRelease(v string) bool {
	m := pep440PreRelease.FindStringSubmatch(v)
	return m != nil && m[2] == "post"
}
