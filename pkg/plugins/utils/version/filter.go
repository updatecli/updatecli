package version

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"

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
)

// SupportedKind holds a list of supported version kind
var SupportedKind []string = []string{
	REGEXVERSIONKIND,
	SEMVERVERSIONKIND,
	LATESTVERSIONKIND,
	REGEXSEMVERVERSIONKIND,
	REGEXTIMEVERSIONKIND,
	TIMEVERSIONKIND,
}

// Filter defines parameters to apply different kind of version matching based on a list of versions
type Filter struct {
	// specifies the version kind such as semver, regex, or latest
	Kind string `yaml:",omitempty"`
	// specifies the version pattern according the version kind
	// for semver, it is a semver constraint
	// for regex, it is a regex pattern
	// for time, it is a date format
	Pattern string `yaml:",omitempty"`
	// strict enforce strict versioning rule.
	// Only used for semantic versioning at this time
	Strict bool `yaml:",omitempty"`
	// specifies the regex pattern, used for regex/semver and regex/time.
	// Output of the first capture group will be used.
	Regex string `yaml:",omitempty"`
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
		case SEMVERVERSIONKIND:
			f.Pattern = "*"
		case LATESTVERSIONKIND:
			f.Pattern = LATESTVERSIONKIND
		default:
			logrus.Warningf("No default pattern provided for kind %q", f.Kind)
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
	return nil
}

// Search returns a value matching pattern
func (f *Filter) Search(versions []string) (Version, error) {
	logrus.Infof("Searching for version matching pattern %q", f.Pattern)

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
			if re.MatchString(v) {
				foundVersion.ParsedVersion = v
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
			found := re.FindStringSubmatch(v)
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

			found := re.FindStringSubmatch(v)

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

// GreaterThanPattern returns a pattern that can be used to find newer version
func (f *Filter) GreaterThanPattern(version string) (string, error) {
	switch f.Kind {
	case LATESTVERSIONKIND:
		return LATESTVERSIONKIND, nil

	case REGEXVERSIONKIND:
		return f.Pattern, nil

	case TIMEVERSIONKIND:
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

			return fmt.Sprintf(">=%d.%d.%d-%s <= %d.%d.%d",
				v.Major(), v.Minor(), v.Patch(), prerelease,
				v.Major(), v.Minor(), v.Patch(),
			), nil

		case "patch":
			v, err := sv.NewVersion(version)
			if err != nil {
				return "", err
			}

			switch v.Prerelease() == "" {
			case true:
				return fmt.Sprintf("%d.%d.x",
					v.Major(),
					v.Minor()), nil
			case false:
				return fmt.Sprintf("%d.%d.x-0",
					v.Major(),
					v.Minor()), nil
			}

		case "minor":
			v, err := sv.NewVersion(version)
			if err != nil {
				return "", err
			}

			switch v.Prerelease() == "" {
			case true:
				return fmt.Sprintf(
					"%d.x",
					v.Major()), nil
			case false:
				return fmt.Sprintf(
					"%d.x.x-0",
					v.Major()), nil
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
				return fmt.Sprintf(
					">=%d",
					v.Major()), nil

			case false:
				return fmt.Sprintf(
					">=%d.x.x-0",
					v.Major()), nil
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
				return ">=" + v.String(), nil
			}

			_, err = sv.NewConstraint(version)
			if err != nil {
				return "", &ErrIncorrectSemVerConstraint{SemVerConstraint: version}
			}
			return version, nil

		default:
			return f.Pattern, nil
		}
	}
	return "", &ErrUnsupportedVersionKind{Kind: f.Kind}
}
