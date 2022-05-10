package helm

import (
	"errors"
	"strings"

	"github.com/sirupsen/logrus"
)

var (
	// ErrWrongConfig is returned in case of a wrong configuration
	ErrWrongConfig = errors.New("wrong helm configuration")
)

//ValidateTarget ensure that target required parameter are set
func (c *Chart) ValidateTarget() error {

	gotErr := false

	if len(c.spec.File) == 0 {
		c.spec.File = "values.yaml"
	}

	if len(c.spec.Name) == 0 {
		gotErr = true
		logrus.Errorf("parameter name required")
	}

	if len(c.spec.Key) == 0 {
		gotErr = true
		logrus.Errorf("parameter key required")
	}

	if len(c.spec.VersionIncrement) == 0 {
		c.spec.VersionIncrement = MINORVERSION
	}

	for _, inc := range strings.Split(c.spec.VersionIncrement, ",") {

		if inc != MAJORVERSION &&
			inc != MINORVERSION &&
			inc != NOINCREMENT &&
			inc != PATCHVERSION &&
			inc != "" {
			gotErr = true
			logrus.Errorf("unrecognized increment rule %q. accepted values are a comma separated list of [major,minor,patch], or 'none' to disable version increment", inc)
		}
	}

	if gotErr {
		return ErrWrongConfig
	}

	return nil
}
