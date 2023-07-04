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
)

var (
	// SupportedKind holds a list of supported version kind
	SupportedKind []string = []string{
		REGEXVERSIONKIND,
		SEMVERVERSIONKIND,
		LATESTVERSIONKIND,
	}
)

// Filter defines parameters to apply different kind of version matching based on a list of versions
type Filter struct {
	// specifies the version kind such as semver, regex, or latest
	Kind string `yaml:",omitempty"`
	// specifies the version pattern according the version kind
	Pattern string `yaml:",omitempty"`
	// strict enforce strict versioning rule. Only used for semantic versioning at this time
	Strict bool `yaml:",omitempty"`
}

// Init returns a new (copy) valid instantiated filter
func (f Filter) Init() (Filter, error) {
	// Set default kind value to "latest"
	if len(f.Kind) == 0 {
		f.Kind = LATESTVERSIONKIND
	}

	// Set default pattern value based on kind
	if f.Kind == LATESTVERSIONKIND && len(f.Pattern) == 0 {
		f.Pattern = LATESTVERSIONKIND
	} else if f.Kind == SEMVERVERSIONKIND && len(f.Pattern) == 0 {
		f.Pattern = "*"
	} else if f.Kind == REGEXVERSIONKIND && len(f.Pattern) == 0 {
		f.Pattern = ".*"
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

	switch f.Kind {
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

	case SEMVERVERSIONKIND:

		switch f.Pattern {
		case "prerelease":
			v, err := sv.NewVersion(version)
			if err != nil {
				return "", err
			}

			return fmt.Sprintf(">=%d.%d.%d-%s <= %d.%d.%d",
				v.Major(), v.Minor(), v.Patch(), v.Prerelease(),
				v.Major(), v.Minor(), v.Patch(),
			), nil

		case "patch":
			v, err := sv.NewVersion(version)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("%d.%d.x",
				v.Major(),
				v.Minor()), nil

		case "minor":
			v, err := sv.NewVersion(version)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf(
				"%d.x",
				v.Major()), nil

		case "minoronly":
			v, err := sv.NewVersion(version)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf(
				"%s || >%d.%d < %d",
				version,
				v.Major(), v.Minor(),
				v.IncMajor().Major(),
			), nil

		case "major":
			v, err := sv.NewVersion(version)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf(
				">=%d",
				v.Major()), nil

		case "majoronly":
			v, err := sv.NewVersion(version)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf(
				"%s || >%d",
				version, v.Major(),
			), nil

		case "", "*":
			v, err := sv.NewVersion(version)
			// If NewVersion do not fails then it means that version contains a valid semantic version
			if err == nil {
				return ">=" + v.String(), nil
			}

			fmt.Println(version)
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
