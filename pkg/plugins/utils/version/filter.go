package version

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/log"

	sv "github.com/Masterminds/semver/v3"
)

const (
	// REGEXVERSIONKIND represents versions as a simple string
	REGEXVERSIONKIND string = "regex"
	// SEMVERVERSIONKIND represents versions as a semantic versioning type
	SEMVERVERSIONKIND string = "semver"
	// LATESTVERSIONKIND specifies that we are looking for the latest version of an array
	LATESTVERSIONKIND string = "latest"
	// REGEXSEMVERVERSIONKIND use regex to extract versions and semantic version to find version
	REGEXSEMVERVERSIONKIND string = "regex/semver"
	// REGEXTIMEVERSIONKIND use regex to extract all date matching format to find version
	REGEXTIMEVERSIONKIND string = "regex/time"
	// TIMEVERSIONKIND represents versions use date format to identify the version
	TIMEVERSIONKIND string = "time"
	// LEXSORTVERSIONKIND uses the go sort package to find the latest version
	LEXVERSIONKIND string = "lex"
	// PEP440VERSIONKIND represents versions as a [pep440](https://peps.python.org/pep-0440/) versioning kind
	PEP440VERSIONKIND string = "pep440"
)

// SupportedKind holds a list of supported version kind
var SupportedKind []string = []string{
	REGEXVERSIONKIND,
	SEMVERVERSIONKIND,
	LATESTVERSIONKIND,
	REGEXSEMVERVERSIONKIND,
	REGEXTIMEVERSIONKIND,
	TIMEVERSIONKIND,
	LEXVERSIONKIND,
	PEP440VERSIONKIND,
}

// ReplaceAll defines a regular expression replacement applied to each version before filtering.
type ReplaceAll struct {
	// "pattern" defines the regular expression matching the text to replace.
	//
	// example:
	//   * pattern: "_"
	//
	Pattern string `yaml:",omitempty"`
	// "replacement" defines the text replacing each match of "pattern".
	//
	// remark:
	//   * capture groups can be referenced with $1, $2, and so on.
	//
	// example:
	//   * replacement: "."
	//
	Replacement string `yaml:",omitempty"`
}

// Filter defines how to select one version from a list of versions.
type Filter struct {
	// "kind" defines the versioning scheme used to select a version.
	//
	// default:
	//   latest
	//
	// remark:
	//   * accepted values are "latest", "semver", "regex", "regex/semver", "time", "regex/time", "lex" and "pep440".
	//   * "latest" returns the last version of the list.
	//   * "lex" sorts the versions lexicographically and returns the last one.
	//   * "pep440" follows https://peps.python.org/pep-0440/
	//
	// example:
	//   * kind: semver
	//
	Kind string `yaml:",omitempty"`
	// "pattern" defines the version pattern, according to "kind".
	//
	// default:
	//   * latest: "latest"
	//   * semver and pep440: "*"
	//   * regex: ".*"
	//   * time and regex/time: "2006-01-02"
	//
	// remark:
	//   * for "latest", "latest" returns the last version, any other value must match a version exactly.
	//   * for "semver" and "regex/semver", it is a semantic versioning constraint.
	//   * for "pep440", it is a pep440 version specifier.
	//   * for "regex", it is a regular expression.
	//   * for "time" and "regex/time", it is a Go date layout.
	//   * ignored by "lex".
	//
	// example:
	//   * pattern: ~1.2
	//   * pattern: ">=1.0.0 <2.0.0"
	//   * pattern: ^v\d+\.\d+\.\d+$
	//
	Pattern string `yaml:",omitempty"`
	// "strict" enforces strict semantic versioning rules when parsing versions.
	//
	// default:
	//   false
	//
	// remark:
	//   * only used by the kinds "semver" and "regex/semver".
	//
	Strict bool `yaml:",omitempty"`
	// "regex" defines the regular expression extracting the version from each entry.
	//
	// remark:
	//   * only used by the kinds "regex/semver" and "regex/time".
	//   * the value of the first capture group is used as the version.
	//
	// example:
	//   * regex: ^v(\d+\.\d+\.\d+)$
	//
	Regex string `yaml:",omitempty"`
	// "replaceall" applies a regular expression replacement to each version before filtering.
	//
	// remark:
	//   * only used by the kinds "regex", "regex/semver" and "regex/time".
	//   * the replacement runs before "pattern" or "regex" is evaluated.
	//
	// example:
	//   ```
	//   replaceall:
	//     pattern: "_"
	//     replacement: "."
	//   ```
	//   turns "curl-8_15_0" into "curl-8.15.0".
	//
	ReplaceAll ReplaceAll `yaml:",omitempty"`
}

// Init returns a new (copy) valid instantiated filter
func (f Filter) Init() (Filter, error) {
	// Set default kind value to "latest"
	if len(f.Kind) == 0 {
		f.Kind = LATESTVERSIONKIND
	}

	// Set default pattern value according to kind
	if len(f.Pattern) == 0 {
		switch f.Kind {
		case TIMEVERSIONKIND, REGEXTIMEVERSIONKIND:
			f.Pattern = "2006-01-02"
		case REGEXVERSIONKIND:
			f.Pattern = ".*"
		case SEMVERVERSIONKIND, PEP440VERSIONKIND:
			f.Pattern = "*"
		case LATESTVERSIONKIND:
			f.Pattern = LATESTVERSIONKIND
		default:
			log.Warningf("No default pattern provided for kind %q", f.Kind)
		}
	}

	return f, f.Validate()
}

// Validate tests if our filter contains valid parameters
func (f Filter) Validate() error {
	ok := false

	for id := range SupportedKind {
		if SupportedKind[id] == f.Kind {
			ok = true
			break
		}
	}
	if !ok {
		return &ErrUnsupportedVersionKind{Kind: f.Kind}
	}

	// Validate ReplaceAll pattern if provided
	if len(f.ReplaceAll.Pattern) > 0 {
		_, err := regexp.Compile(f.ReplaceAll.Pattern)
		if err != nil {
			return fmt.Errorf("invalid replaceAll pattern %q: %w", f.ReplaceAll.Pattern, err)
		}
	}

	return nil
}

// applyReplaceAll applies the ReplaceAll transformation to a version string if configured
func (f *Filter) applyReplaceAll(version string) (string, error) {
	if len(f.ReplaceAll.Pattern) == 0 {
		return version, nil
	}

	re, err := regexp.Compile(f.ReplaceAll.Pattern)
	if err != nil {
		return "", fmt.Errorf("compiling replaceAll pattern %q: %w", f.ReplaceAll.Pattern, err)
	}

	return re.ReplaceAllString(version, f.ReplaceAll.Replacement), nil
}

// Search returns a value matching pattern
func (f *Filter) Search(versions []string) (Version, error) {
	log.Infof("Searching for version matching pattern %q", f.Pattern)

	return f.search(versions)
}

// Matches returns true if one of versions matches the filter. Unlike Search, it doesn't log
// the pattern it searches for.
func (f *Filter) Matches(versions []string) bool {
	_, err := f.search(versions)
	return err == nil
}

// search returns the version matching the filter among versions.
func (f *Filter) search(versions []string) (Version, error) {
	foundVersion := Version{}

	if len(versions) == 0 {
		return foundVersion, ErrNoVersionFound
	}

	switch f.Kind {
	case TIMEVERSIONKIND:
		d := Time{
			layout: f.Pattern,
		}

		mapVersions := make(map[string]string)
		for _, v := range versions {
			mapVersions[v] = v
		}

		err := d.Search(mapVersions)
		if err != nil {
			return foundVersion, err
		}

		return d.FoundVersion, nil

	case LATESTVERSIONKIND:
		if f.Pattern == LATESTVERSIONKIND {
			foundVersion.ParsedVersion = versions[len(versions)-1]
			foundVersion.OriginalVersion = foundVersion.ParsedVersion
			return foundVersion, nil
		}
		// Search for simple text matching
		for i := len(versions) - 1; i >= 0; i-- {
			if strings.Compare(f.Pattern, versions[i]) == 0 {
				foundVersion.ParsedVersion = versions[i]
				foundVersion.OriginalVersion = versions[i]
				return foundVersion, nil
			}
		}
	case REGEXVERSIONKIND:
		re, err := regexp.Compile(f.Pattern)
		if err != nil {
			return foundVersion, err
		}

		// Parse version in by date publishing
		// Oldest version appears first in array
		for i := len(versions) - 1; i >= 0; i-- {
			v := versions[i]
			// Apply replaceAll transformation if configured
			transformedV, err := f.applyReplaceAll(v)
			if err != nil {
				return foundVersion, err
			}
			if re.MatchString(transformedV) {
				foundVersion.ParsedVersion = transformedV
				foundVersion.OriginalVersion = v
				return foundVersion, nil
			}
		}
	case SEMVERVERSIONKIND:
		s := Semver{
			Constraint: f.Pattern,
			Strict:     f.Strict,
		}

		err := s.Search(versions)
		if err != nil {
			return foundVersion, err
		}

		return s.FoundVersion, nil
	case REGEXSEMVERVERSIONKIND:
		re, err := regexp.Compile(f.Regex)
		if err != nil {
			return foundVersion, err
		}

		// Create a slice of versions using regex pattern
		var parsedVersions []string
		versionLookup := make(map[string]string)
		for i := 0; i < len(versions); i++ {
			v := versions[i]
			// Apply replaceAll transformation if configured
			transformedV, err := f.applyReplaceAll(v)
			if err != nil {
				return foundVersion, err
			}
			found := re.FindStringSubmatch(transformedV)
			if len(found) > 1 {
				parsedVersions = append(parsedVersions, found[1])
				versionLookup[found[1]] = v
			}

		}

		s := Semver{
			Constraint: f.Pattern,
			Strict:     f.Strict,
		}

		err = s.Search(parsedVersions)
		if err != nil {
			return foundVersion, err
		}

		if originalVersion, ok := versionLookup[s.FoundVersion.OriginalVersion]; ok {
			s.FoundVersion.OriginalVersion = originalVersion
		}

		return s.FoundVersion, nil

	case REGEXTIMEVERSIONKIND:
		re, err := regexp.Compile(f.Regex)
		if err != nil {
			return foundVersion, err
		}

		// Create a slice of versions using regex pattern
		parsedVersions := make(map[string]string)
		for i := 0; i < len(versions); i++ {
			v := versions[i]
			// Apply replaceAll transformation if configured
			transformedV, err := f.applyReplaceAll(v)
			if err != nil {
				return foundVersion, err
			}

			found := re.FindStringSubmatch(transformedV)

			if len(found) > 1 {
				parsedVersions[found[1]] = v
			}

		}

		d := Time{
			layout: f.Pattern,
		}

		if err = d.Search(parsedVersions); err != nil {
			return foundVersion, err
		}

		return d.FoundVersion, nil

	case LEXVERSIONKIND:
		slices.Sort(versions)
		foundVersion.ParsedVersion = versions[len(versions)-1]
		foundVersion.OriginalVersion = foundVersion.ParsedVersion
		return foundVersion, nil

	case PEP440VERSIONKIND:
		p := Pep440{
			Constraint: f.Pattern,
		}

		err := p.Search(versions)
		if err != nil {
			return foundVersion, err
		}

		return p.FoundVersion, nil

	default:
		return foundVersion, &ErrUnsupportedVersionKindPattern{Pattern: f.Pattern, Kind: f.Kind}
	}

	return foundVersion, &ErrNoVersionFoundForPattern{Pattern: f.Pattern}
}

// IsZero return true if filter is not initialized
func (f Filter) IsZero() bool {
	var empty Filter
	return empty == f
}

// BoundPattern returns a semver pattern that also rejects any version lower than lowerBound.
// Each group of the pattern, separated by "||", gets lowerBound as minimum unless it already
// rejects every version lower than it. The pattern is returned unchanged when lowerBound isn't a version or the
// pattern isn't a semver constraint, such as "latest".
//
// A group only accepts prereleases when one of its comparators is a prerelease. So that the
// bound doesn't make a group accept prereleases, a group without any, such as "^1.2.0", is
// bounded at the release of a prerelease lowerBound: "1.5.0" for "1.5.0-beta.1".
func BoundPattern(pattern, lowerBound string) string {
	return boundPattern(pattern, lowerBound, false)
}

// BoundPatternAcceptingPrereleases returns the same pattern as BoundPattern, except that every
// group is bounded at lowerBound itself. For a prerelease lowerBound, such as a Go pseudo
// version, every group then also accepts the prereleases newer than it.
func BoundPatternAcceptingPrereleases(pattern, lowerBound string) string {
	return boundPattern(pattern, lowerBound, true)
}

func boundPattern(pattern, lowerBound string, acceptPrereleases bool) string {
	v, err := sv.NewVersion(lowerBound)
	if err != nil {
		return pattern
	}

	// groupBound returns the bound of a group, which only accepts prereleases when it already does
	groupBound := func(group string) *sv.Version {
		if acceptPrereleases || v.Prerelease() == "" || prereleaseComparatorRegex.MatchString(buildMetadataRegex.ReplaceAllString(group, "")) {
			return v
		}
		return sv.New(v.Major(), v.Minor(), v.Patch(), "", "")
	}

	/*
		A pattern accepting any version only needs the bound. ">=0.0.0-0" also accepts
		prereleases, which only a prerelease bound keeps, so a release bound is added to it.
	*/
	if pattern == "" || pattern == "*" || (pattern == anyPrereleasePattern && v.Prerelease() != "") {
		return ">=" + groupBound(pattern).String()
	}

	if _, err := sv.NewConstraint(pattern); err != nil {
		return pattern
	}

	groups := strings.Split(pattern, "||")
	for i, group := range groups {
		group = strings.TrimSpace(group)
		bound := groupBound(group)
		if !hasLowerBound(group, bound) {
			group = fmt.Sprintf(">=%s, %s", bound.String(), group)
		}
		groups[i] = group
	}

	return strings.Join(groups, " || ")
}

// anyPrereleasePattern is a semver pattern accepting any version, prereleases included.
const anyPrereleasePattern = ">=0.0.0-0"

// prereleaseComparatorRegex matches a comparator whose version is a prerelease, such as
// "1.2.x-0" or ">=1.5.0-beta.1", but not a hyphen range such as "1.2 - 2.0".
var prereleaseComparatorRegex = regexp.MustCompile(`[0-9xX*]-[0-9A-Za-z]`)

// buildMetadataRegex matches the build metadata of a comparator, such as "+2026-01" in
// "^1.2.0+2026-01", whose hyphens aren't prereleases.
var buildMetadataRegex = regexp.MustCompile(`\+[0-9A-Za-z.-]*`)

// UnboundedPattern returns a semver pattern accepting any version, such as one lower than
// currentVersion. For a prerelease, it also accepts prereleases, which "*" doesn't, so that
// a prerelease in use isn't replaced by an older stable release.
func UnboundedPattern(currentVersion string) string {
	if v, err := sv.NewVersion(currentVersion); err == nil && v.Prerelease() != "" {
		return anyPrereleasePattern
	}
	return "*"
}

// lowerBoundOperators lists the operators of a comparator rejecting any version lower than its own.
var lowerBoundOperators = []string{"", "=", ">", ">=", "^", "~", "~>"}

// operatorSpaceRegex matches the spaces a constraint allows between an operator and its version, as in "> 1.2.5".
var operatorSpaceRegex = regexp.MustCompile(`([<>=!~^]+)\s+`)

// hasLowerBound returns true if a group of a semver constraint, without "||", already rejects
// every version lower than bound.
//
// A group only accepts prereleases when one of its comparators is a prerelease, so for a
// prerelease bound, such as a Go pseudo version, only a prerelease comparator counts: the
// bound added otherwise also makes the group accept prereleases.
func hasLowerBound(group string, bound *sv.Version) bool {
	// Only the first version of a hyphen range, such as "1.2.5 - 2.0.0", is a lower bound
	if first, _, found := strings.Cut(group, " - "); found {
		group = first
	}

	group = operatorSpaceRegex.ReplaceAllString(group, "$1")

	for _, comparator := range strings.FieldsFunc(group, func(r rune) bool { return r == ',' || r == ' ' }) {
		operatorEnd := strings.IndexFunc(comparator, func(r rune) bool { return !strings.ContainsRune("<>=!~^", r) })
		if operatorEnd < 0 || !slices.Contains(lowerBoundOperators, comparator[:operatorEnd]) {
			continue
		}

		v, err := sv.NewVersion(comparator[operatorEnd:])
		if err != nil {
			continue
		}

		if bound.Prerelease() != "" && v.Prerelease() == "" {
			continue
		}

		if !v.LessThan(bound) {
			return true
		}
	}

	return false
}

// lowerBoundMode defines how a pattern built from a version handles the versions lower than it.
type lowerBoundMode int

const (
	// defaultLowerBound only bounds the "*" pattern at the given version
	defaultLowerBound lowerBoundMode = iota
	// withLowerBound bounds every semver pattern at the given version
	withLowerBound
	// withoutLowerBound bounds no pattern
	withoutLowerBound
)

// GreaterThanPattern returns a pattern that can be used to find newer version
func (f *Filter) GreaterThanPattern(version string) (string, error) {
	return f.greaterThanPattern(version, defaultLowerBound)
}

// GreaterThanPatternWithoutDowngrade returns the same pattern as GreaterThanPattern, except
// that a semver pattern never accepts a version lower than the given one. A source whose
// newest versions are filtered out, such as by an age filter, then finds nothing rather than
// falling back to an older version. A custom semver constraint is bounded the same way, unless
// it already rejects every version lower than the given one.
func (f *Filter) GreaterThanPatternWithoutDowngrade(version string) (string, error) {
	return f.greaterThanPattern(version, withLowerBound)
}

// GreaterThanPatternAllowingDowngrade returns the same pattern as GreaterThanPattern,
// without any lower bound. A source whose newest versions are filtered out, such as by
// an age filter, can then fall back to a version older than the given one.
func (f *Filter) GreaterThanPatternAllowingDowngrade(version string) (string, error) {
	return f.greaterThanPattern(version, withoutLowerBound)
}

// boundAt prepends ">=v" to pattern when mode bounds every pattern.
func boundAt(pattern string, v *sv.Version, mode lowerBoundMode) string {
	if mode != withLowerBound {
		return pattern
	}
	return fmt.Sprintf(">=%s, %s", v.String(), pattern)
}

func (f *Filter) greaterThanPattern(version string, mode lowerBoundMode) (string, error) {
	switch f.Kind {
	case LATESTVERSIONKIND:
		return LATESTVERSIONKIND, nil

	case REGEXVERSIONKIND:
		return f.Pattern, nil

	case TIMEVERSIONKIND:
		return f.Pattern, nil

	case REGEXSEMVERVERSIONKIND:
		return f.Pattern, nil

	case REGEXTIMEVERSIONKIND:
		return f.Pattern, nil

	case SEMVERVERSIONKIND:

		switch f.Pattern {
		case "prerelease":
			v, err := sv.NewVersion(version)
			if err != nil {
				return "", err
			}

			prerelease := v.Prerelease()
			if prerelease == "" {
				prerelease = "0"
			}

			// For a release, ">=1.2.3-0" also accepts the prereleases lower than it
			return boundAt(fmt.Sprintf(">=%d.%d.%d-%s <= %d.%d.%d",
				v.Major(), v.Minor(), v.Patch(), prerelease,
				v.Major(), v.Minor(), v.Patch(),
			), v, mode), nil

		case "patch":
			v, err := sv.NewVersion(version)
			if err != nil {
				return "", err
			}

			switch v.Prerelease() == "" {
			case true:
				return boundAt(fmt.Sprintf("%d.%d.x",
					v.Major(),
					v.Minor()), v, mode), nil
			case false:
				return boundAt(fmt.Sprintf("%d.%d.x-0",
					v.Major(),
					v.Minor()), v, mode), nil
			}

		case "minor":
			v, err := sv.NewVersion(version)
			if err != nil {
				return "", err
			}

			switch v.Prerelease() == "" {
			case true:
				return boundAt(fmt.Sprintf("%d.x", v.Major()), v, mode), nil
			case false:
				return boundAt(fmt.Sprintf("%d.x.x-0", v.Major()), v, mode), nil
			}

		case "minoronly":
			v, err := sv.NewVersion(version)
			if err != nil {
				return "", err
			}

			switch v.Prerelease() == "" {
			case true:
				return fmt.Sprintf(
					"%s || >%d.%d < %d",
					version,
					v.Major(), v.Minor(),
					v.IncMajor().Major(),
				), nil
			case false:
				return fmt.Sprintf(
					"%s || >%d.%d.x-0 < %d",
					version,
					v.Major(), v.Minor(),
					v.IncMajor().Major(),
				), nil
			}

		case "major":
			v, err := sv.NewVersion(version)
			if err != nil {
				return "", err
			}
			switch v.Prerelease() == "" {
			case true:
				if mode != withLowerBound {
					return fmt.Sprintf(">=%d", v.Major()), nil
				}
				return ">=" + v.String(), nil

			case false:
				return boundAt(fmt.Sprintf(">=%d.x.x-0", v.Major()), v, mode), nil
			}

		case "majoronly":
			v, err := sv.NewVersion(version)
			if err != nil {
				return "", err
			}
			switch v.Prerelease() == "" {
			case true:
				return fmt.Sprintf(
					"%s || >%d",
					version, v.Major(),
				), nil
			case false:
				return fmt.Sprintf(
					"%s || >%s",
					version, version,
				), nil
			}

		case "", "*":
			v, err := sv.NewVersion(version)
			// If NewVersion do not fails then it means that version contains a valid semantic version
			if err == nil {
				if mode == withoutLowerBound {
					return UnboundedPattern(version), nil
				}
				return ">=" + v.String(), nil
			}

			_, err = sv.NewConstraint(version)
			if err != nil {
				return "", &ErrIncorrectSemVerConstraint{SemVerConstraint: version}
			}
			return version, nil

		default:
			// A custom constraint, such as "~1.5", may also accept versions lower than the
			// current one
			if mode != withLowerBound {
				return f.Pattern, nil
			}
			return BoundPattern(f.Pattern, version), nil
		}

	case PEP440VERSIONKIND:
		return f.Pattern, nil
	}
	return "", &ErrUnsupportedVersionKind{Kind: f.Kind}
}
