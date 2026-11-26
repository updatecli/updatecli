package version

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	pep440 "github.com/aquasecurity/go-pep440-version"
	"github.com/updatecli/updatecli/pkg/core/log"
)

var (
	// pep440PreSegmentRegex matches the pre-release segment of a normalized PEP 440 version, such as "b1" in "2.0b1".
	pep440PreSegmentRegex = regexp.MustCompile(`\d(a|b|rc)\d+`)
	// pep440DevSegmentRegex matches the dev segment ending a normalized PEP 440 version, such as ".dev2" in "1.0.post1.dev2".
	pep440DevSegmentRegex = regexp.MustCompile(`\.dev\d+$`)
)

// Pep440 filters versions using PEP 440 specifier syntax (e.g. ">=1.0,<2.0").
type Pep440 struct {
	Constraint   string
	versions     []pep440.Version
	FoundVersion Version
}

// Init parses each version string and retains those that are valid PEP 440 versions.
func (p *Pep440) Init(versions []string) error {
	for _, raw := range versions {
		v, err := pep440.Parse(raw)
		if err != nil {
			log.Debugf("Skipping %q because %s", raw, err)
			continue
		}
		p.versions = append(p.versions, v)
	}

	if len(p.versions) > 0 {
		return nil
	}

	return ErrNoValidPep440VersionFound
}

// Sort orders versions descending so the newest is first.
func (p *Pep440) Sort() {
	sort.Slice(p.versions, func(i, j int) bool {
		return p.versions[i].Compare(p.versions[j]) > 0
	})
}

// Search returns the highest version satisfying the constraint, or the highest
// version overall when Constraint is empty or "*".
func (p *Pep440) Search(versions []string) error {
	if len(versions) == 0 {
		return ErrNoVersionsFound
	}

	if err := p.Init(versions); err != nil {
		log.Error(err)
		return err
	}

	p.Sort()

	if p.Constraint == "" || p.Constraint == "*" {
		// Prefer stable versions over pre-releases.
		for _, v := range p.versions {
			if !v.IsPreRelease() {
				p.FoundVersion.ParsedVersion = v.String()
				p.FoundVersion.OriginalVersion = v.Original()
				return nil
			}
		}
		// Fall back to highest pre-release if no stable version exists.
		p.FoundVersion.ParsedVersion = p.versions[0].String()
		p.FoundVersion.OriginalVersion = p.versions[0].Original()
		return nil
	}

	specifiers, err := pep440.NewSpecifiers(p.Constraint)
	if err != nil {
		// Normalize the xerrors-typed error from the pep440 library to a standard error type.
		return fmt.Errorf("%s", err)
	}

	// A pre-release only qualifies once no stable version satisfies the constraint.
	for _, v := range p.versions {
		if specifiers.Check(v) && !v.IsPreRelease() {
			p.FoundVersion.ParsedVersion = v.String()
			p.FoundVersion.OriginalVersion = v.Original()
			return nil
		}
	}

	for _, v := range p.versions {
		if specifiers.Check(v) {
			p.FoundVersion.ParsedVersion = v.String()
			p.FoundVersion.OriginalVersion = v.Original()
			return nil
		}
	}

	return ErrNoVersionFound
}

// boundPEP440Pattern returns pattern without any version lower than lowerBound, each group
// of an "||" pattern getting its own bound. "*" becomes the bound alone.
//
// A specifier can't name a local version, so "2.1.0+cpu" is bounded at "2.1.0". A pre-release
// lowerBound, such as "2.0b1", bounds a group at its release, "2.0", unless the group already
// names a pre-release: in a PEP 440 specifier, naming one makes every pre-release match.
//
// It returns an error when lowerBound isn't a PEP 440 version or pattern isn't a valid
// PEP 440 specifier, as a pattern then can't be bounded.
func boundPEP440Pattern(pattern, lowerBound string) (string, error) {
	v, err := pep440.Parse(lowerBound)
	if err != nil {
		return "", fmt.Errorf("%q is not a PEP 440 version: %w", lowerBound, err)
	}

	// groupBound returns the bound of a group, which only accepts pre-releases when it already does
	groupBound := func(group string) string {
		if v.IsPreRelease() && !namesPEP440PreRelease(group) {
			return ">=" + pep440ReleaseAbove(v)
		}
		return ">=" + v.Public()
	}

	pattern = strings.TrimSpace(pattern)
	if pattern == "" || pattern == "*" {
		return groupBound(""), nil
	}

	if _, err := pep440.NewSpecifiers(pattern); err != nil {
		return "", fmt.Errorf("%q is not a PEP 440 specifier: %w", pattern, err)
	}

	groups := strings.Split(pattern, "||")
	for i, group := range groups {
		group = strings.TrimSpace(group)
		groups[i] = groupBound(group) + ", " + group
	}
	bounded := strings.Join(groups, " || ")

	if _, err := pep440.NewSpecifiers(bounded); err != nil {
		return "", fmt.Errorf("bounding %q at %q: %w", pattern, lowerBound, err)
	}
	return bounded, nil
}

// pep440ReleaseAbove returns the lowest version above the pre-release v that isn't a pre-release:
// its release, such as "2.0" for "2.0b1" or "2.0.dev3", or, for a dev release of a post release
// such as "1.0.post1.dev2", that post release, as its release "1.0" is lower than it.
func pep440ReleaseAbove(v pep440.Version) string {
	public := v.Public()
	if v.IsPostRelease() && !pep440PreSegmentRegex.MatchString(public) {
		return pep440DevSegmentRegex.ReplaceAllString(public, "")
	}
	return v.BaseVersion()
}

// namesPEP440PreRelease returns true if a specifier group, without "||", names a pre-release,
// such as ">=2.0b1, <3".
func namesPEP440PreRelease(group string) bool {
	for _, specifier := range strings.Split(group, ",") {
		ver := strings.TrimLeft(strings.TrimSpace(specifier), "<>=!~ ")
		ver = strings.TrimSuffix(ver, ".*")
		if v, err := pep440.Parse(ver); err == nil && v.IsPreRelease() {
			return true
		}
	}
	return false
}
