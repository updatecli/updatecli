package version

import (
	"fmt"
	"regexp"

	"github.com/olblak/updateCli/pkg/plugins/version/semver"
	"github.com/sirupsen/logrus"
)

// Version hold version information
type Version struct {
	Kind    string
	Pattern string
}

const (
	// TEXTVERSIONKIND represent versions as a simple string
	TEXTVERSIONKIND string = "text"
	// SEMVERVERSIONKIND represent versions as a semantic versionning type
	SEMVERVERSIONKIND string = "semver"
)

var (
	// SupportedKind holds a list of supported version kind
	SupportedKind []string = []string{
		TEXTVERSIONKIND,
		SEMVERVERSIONKIND,
	}
)

// Validate tests if we are analysing a valid version type
func (v *Version) Validate() error {

	if len(v.Kind) == 0 {
		v.Kind = TEXTVERSIONKIND
	}

	ok := false
	for id := range SupportedKind {
		if SupportedKind[id] == v.Kind {
			ok = true
			break
		}
	}
	if !ok {
		return fmt.Errorf("Unsupported version kind %q", v.Kind)
	}
	return nil
}

// Search returns a value matching pattern
func (v *Version) Search(versions []string) (version string, err error) {

	logrus.Infof("Searching for version matching pattern %q", v.Pattern)

	switch v.Kind {
	case TEXTVERSIONKIND:
		if v.Pattern == "latest" {
			version = versions[len(versions)-1]
		} else {
			re, err := regexp.Compile(v.Pattern)
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
		}
	case SEMVERVERSIONKIND:
		s := semver.Semver{
			Constraint: v.Pattern,
		}

		version, err = s.Searcher(versions)
		if err != nil {
			logrus.Error(err)
			return version, err
		}
	default:
		return version, fmt.Errorf("Unsupported version kind %q with pattern %q", v.Kind, v.Pattern)

	}

	return version, err
}
