package jenkins

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// Spec defines a specification for a "jenkins" resource
// parsed from an updatecli manifest file
type Spec struct {
	// [s][c] Defines the release name. It accepts "stable" or "weekly"
	Release string `yaml:",omitempty"`
	// [s][c] Defines a specific release version (condition only)
	Version string `yaml:",omitempty"`
}

func (s Spec) Atomic() Spec {
	return Spec{
		Release: s.Release,
		Version: s.Version,
	}
}

// Validate run some validation on the Jenkins struct
func (s Spec) Validate() (err error) {
	if len(s.Release) == 0 && len(s.Version) == 0 {
		logrus.Debugln("Jenkins release type not defined, default set to stable")
		s.Release = "stable"
	} else if len(s.Release) == 0 && len(s.Version) != 0 {
		s.Release, err = ReleaseType(s.Version)
		logrus.Debugf("Jenkins release type not defined, guessing based on Version %s", s.Version)
		if err != nil {
			return err
		}
	}

	if s.Release != WEEKLY &&
		s.Release != STABLE {
		return fmt.Errorf("wrong Jenkins release type '%s', accepted values ['%s','%s']",
			s.Release, WEEKLY, STABLE)

	}
	return nil
}
