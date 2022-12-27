package helm

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	// ErrWrongConfig is the error message for a wrong helm configuration
	ErrWrongConfig = errors.New("wrong helm configuration")
)

// ValidateTarget validates target struct fields.
func (c *Chart) ValidateTarget() error {

	var errs []error

	errs = c.validateVersionInc()

	if len(c.spec.File) == 0 {
		c.spec.File = "values.yaml"
	}

	if len(c.spec.Name) == 0 {
		errs = append(errs, errors.New("parameter name required"))
	}

	if len(c.spec.Key) == 0 {
		errs = append(errs, errors.New("parameter key required"))
	}

	if len(errs) > 0 {
		for _, e := range errs {
			logrus.Errorln(e)
		}
		return ErrWrongConfig
	}

	return nil
}

// validateVersionInc validates that the version increment settings is correctly set.
func (c *Chart) validateVersionInc() []error {
	var errs []error

	// Set default value to minor
	if len(c.spec.VersionIncrement) == 0 {
		c.spec.VersionIncrement = MINORVERSION
		return nil
	}

	// Checks one many time a string appears in a list of string
	isNotDuplicated := func(rules []string, rule string) error {
		counter := 0
		for _, r := range rules {
			if r == rule {
				counter++
			}
		}

		if counter > 0 {
			return fmt.Errorf("rule %q appears multiple time in %v", rule, c.spec.VersionIncrement)
		}
		return nil
	}

	acceptedRules := []string{}
	versionIncrement := strings.Split(c.spec.VersionIncrement, ",")

	for _, inc := range versionIncrement {

		switch inc {
		case MAJORVERSION:
			err := isNotDuplicated(acceptedRules, inc)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			acceptedRules = append(acceptedRules, inc)

		case MINORVERSION:
			err := isNotDuplicated(acceptedRules, inc)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			acceptedRules = append(acceptedRules, inc)

		case NOINCREMENT:
			if len(versionIncrement) > 1 {
				errs = append(
					errs, fmt.Errorf("rule %q is not compatible with others from %q",
						inc, c.spec.VersionIncrement))
			}
		case PATCHVERSION:
			err := isNotDuplicated(acceptedRules, inc)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			acceptedRules = append(acceptedRules, inc)
		default:
			errs = append(errs, fmt.Errorf("unrecognized increment rule %q, accepted values are a comma separated list of [major,minor,patch], or 'none' to disable version increment", inc))
		}
	}

	return errs

}
