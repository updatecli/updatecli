package semver

import (
	"errors"
	"fmt"
	"sort"

	sv "github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// Semver is an interface in front the masterminds/semver used across the updatecli project
type Semver struct {
	Constraint string
	versions   []*sv.Version
}

var (
	// ErrNoVersionFound return a error when no version couldn't be found
	ErrNoVersionFound error = errors.New("No version found")
	// ErrNoVersionsFound return a error when the versions list is empty
	ErrNoVersionsFound error = errors.New("Versions list empty")
)

// Init creates a new semver object
func (s *Semver) Init(versions []string) error {

	for _, version := range versions {
		v, err := sv.NewVersion(version)
		if err != nil {
			logrus.Debugf("Skipping %q because %s, skipping", version, err)
		} else {
			s.versions = append(s.versions, v)
		}
	}

	if len(s.versions) > 0 {
		return nil
	}

	return fmt.Errorf("No valid semantic version found")
}

// Sort re-order a list of versions with the newest version first
func (s *Semver) Sort() {
	sort.Sort(sort.Reverse(sv.Collection(s.versions)))
}

// Search returns the version matching pattern from a sorted list.
func (s *Semver) Search(versions []string) (version string, err error) {
	// We need to be sure that at least one version exist
	if len(versions) == 0 {
		return "", ErrNoVersionsFound

	}
	err = s.Init(versions)
	if err != nil {
		logrus.Error(err)
		return version, err
	}

	s.Sort()

	if len(s.Constraint) == 0 {
		return s.versions[0].String(), err
	}

	c, err := sv.NewConstraint(s.Constraint)
	if err != nil {
		return version, err
	}

	for _, v := range s.versions {

		if c.Check(v) {
			version = v.String()
			break
		}
	}
	if len(version) == 0 {
		return "", ErrNoVersionFound
	}

	return version, err
}
