package osv

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

// maxCandidateChecks bounds the number of candidate versions queried to find a fixed version.
const maxCandidateChecks = 20

var (
	// ErrNoFixedVersion is returned when no published version fixes every known vulnerability yet.
	// It reports a waiting state rather than a lookup failure, so the source is skipped instead of failing.
	ErrNoFixedVersion = errors.New("no version fixing all known vulnerabilities")
	// ErrUnsupportedEcosystem is returned when the plugin cannot order the versions of an ecosystem.
	ErrUnsupportedEcosystem = errors.New("fixed version not supported for ecosystem")
)

// fixedVersion returns the lowest version, from the spec version, without known vulnerabilities.
// OSV decides whether each candidate is affected, so version ranges are never evaluated locally.
func (o *Osv) fixedVersion(ctx context.Context) (string, error) {
	scheme, err := versionSchemeOf(o.ecosystem)
	if err != nil {
		return "", fmt.Errorf("%w, use key %q or an osv condition instead", err, KeyIDs)
	}

	current := o.spec.Version
	if _, err := scheme.Compare(current, current); err != nil {
		return "", fmt.Errorf("parsing version %q of %s: %w", current, o.packageLabel(), err)
	}

	groups, err := o.knownVulnerabilities(ctx, current)
	if err != nil {
		return "", err
	}
	if len(groups) == 0 {
		return current, nil
	}

	/*
		OSV records the first release fixing a vulnerability, which may be a release candidate.
		Pre-release fixes are only considered from a pre-release version, so that a stable
		version is never updated to a pre-release.
	*/
	allowPrerelease := scheme.IsPrerelease(current)

	candidates := map[string]struct{}{}
	tested := map[string]struct{}{}

	bound, err := addCandidates(candidates, groups, current, scheme, allowPrerelease)
	if err != nil {
		return "", err
	}

	for range maxCandidateChecks {
		next := lowestUntested(candidates, tested, bound, scheme)
		if next == "" {
			break
		}
		tested[next] = struct{}{}

		groups, err = o.knownVulnerabilities(ctx, next)
		if err != nil {
			return "", err
		}
		if len(groups) == 0 {
			return o.withVersionPrefix(next), nil
		}

		logrus.Debugf("version %q of %s is still affected by %d known vulnerabilities", next, o.packageLabel(), len(groups))

		// A candidate can be affected by vulnerabilities that did not affect the current version.
		bound, err = addCandidates(candidates, groups, next, scheme, allowPrerelease)
		if err != nil {
			return "", err
		}
	}

	return "", fmt.Errorf("%w after checking %d candidate versions", ErrNoFixedVersion, len(tested))
}

// withVersionPrefix keeps the "v" prefix of Go versions, which OSV records omit.
func (o *Osv) withVersionPrefix(version string) string {
	if o.ecosystem == ecosystemGo && strings.HasPrefix(o.spec.Version, "v") && !strings.HasPrefix(version, "v") {
		return "v" + version
	}
	return version
}

// addCandidates adds every fixed version newer than from to candidates and returns the lowest
// version that can fix all groups, which is the highest of each group's lowest fix.
func addCandidates(candidates map[string]struct{}, groups []vulnGroup, from string, scheme versionScheme, allowPrerelease bool) (string, error) {
	bound := ""

	for _, group := range groups {
		lowest := ""
		var prereleases []string

		for _, fixed := range group.Fixed {
			result, err := scheme.Compare(fixed, from)
			if err != nil {
				logrus.Debugf("ignoring fixed version %q of %q: %s", fixed, group.ID, err)
				continue
			}
			if result <= 0 {
				continue
			}

			if !allowPrerelease && scheme.IsPrerelease(fixed) {
				prereleases = append(prereleases, fixed)
				continue
			}

			candidates[fixed] = struct{}{}
			if lowest == "" || mustCompare(scheme, fixed, lowest) < 0 {
				lowest = fixed
			}
		}

		if lowest == "" {
			if len(prereleases) > 0 {
				return "", fmt.Errorf("%w, %q is only fixed in pre-release %s", ErrNoFixedVersion, group.ID, strings.Join(prereleases, ", "))
			}
			return "", fmt.Errorf("%w, %q has no fix newer than %q", ErrNoFixedVersion, group.ID, from)
		}

		if bound == "" || mustCompare(scheme, lowest, bound) > 0 {
			bound = lowest
		}
	}

	return bound, nil
}

// lowestUntested returns the lowest candidate at or above bound that was not queried yet.
func lowestUntested(candidates, tested map[string]struct{}, bound string, scheme versionScheme) string {
	lowest := ""

	for candidate := range candidates {
		if _, ok := tested[candidate]; ok {
			continue
		}
		if mustCompare(scheme, candidate, bound) < 0 {
			continue
		}

		if lowest == "" {
			lowest = candidate
			continue
		}
		// Break ties between equivalent spellings, such as "1.0" and "1.0.0", deterministically.
		if result := mustCompare(scheme, candidate, lowest); result < 0 || (result == 0 && candidate < lowest) {
			lowest = candidate
		}
	}

	return lowest
}
