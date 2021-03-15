package semver

import (
	"fmt"
	"sort"

	sv "github.com/Masterminds/semver/v3"
)

// Semver is an interface in front the masterminds/semver used across the updatecli project
type Semver struct {
	Constraint string
	versions   []*sv.Version
}

// Init creates a new semver object
func (s *Semver) Init(versions []string) error {

	vs := make([]*sv.Version, len(versions))
	for i, version := range versions {
		v, err := sv.NewVersion(version)
		if err != nil {
			return fmt.Errorf("Error parsing version: %s", err)
		}

		vs[i] = v
	}

	s.versions = vs

	return nil
}

// Sort re-order a list of versions with the newest first
func (s *Semver) Sort() {
	sort.Sort(sort.Reverse(sv.Collection(s.versions)))
}

// GetLatestVersion return the latest version matching pattern from a sorted list.
func (s *Semver) GetLatestVersion() (version string, err error) {
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
			return version, err
		}
	}

	return version, err
}
