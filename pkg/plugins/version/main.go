package version

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/olblak/updateCli/pkg/plugins/version/semver"
	"github.com/sirupsen/logrus"
)

// Filter defines parameters to apply different kind of version matching based on a list of versions
type Filter struct {
	Kind    string
	Pattern string
}

const (
	// REGEXVERSIONKIND represents versions as a simple string
	REGEXVERSIONKIND string = "regex"
	// SEMVERVERSIONKIND represents versions as a semantic versionning type
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

// Validate tests if our filter contains valid parameters
func (f *Filter) Validate() error {

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

	ok := false

	for id := range SupportedKind {
		if SupportedKind[id] == f.Kind {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("Unsupported version kind %q", f.Kind)
	}
	return nil
}

// Search returns a value matching pattern
func (f *Filter) Search(versions []string) (version string, err error) {

	logrus.Infof("Searching for version matching pattern %q", f.Pattern)

	switch f.Kind {
	case LATESTVERSIONKIND:
		if f.Pattern == LATESTVERSIONKIND {
			version = versions[len(versions)-1]
		}
		// Search for simple text matching
		for i := len(versions) - 1; i >= 0; i-- {
			if strings.Compare(f.Pattern, versions[i]) == 0 {
				version = versions[i]
				break
			}
		}
	case REGEXVERSIONKIND:
		re, err := regexp.Compile(f.Pattern)
		if err != nil {
			return "", err
		}

		// Parse version in by date publishing
		// Oldest version appears first in array
		for i := len(versions) - 1; i >= 0; i-- {
			v := versions[i]
			if re.Match([]byte(v)) {
				version = v
				break
			}
		}
	case SEMVERVERSIONKIND:
		s := semver.Semver{
			Constraint: f.Pattern,
		}

		version, err = s.Search(versions)
		if err != nil {
			logrus.Error(err)
			return version, err
		}
	default:
		return version, fmt.Errorf("Unsupported version kind %q with pattern %q", f.Kind, f.Pattern)

	}

	return version, err
}
