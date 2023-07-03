package version

import (
	"sort"

	sv "github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

// Semver is an interface in front the masterminds/semver used across the updatecli project
type Semver struct {
	Constraint   string
	versions     []*sv.Version
	FoundVersion Version
	Strict       bool
}

// Init creates a new semver object
func (s *Semver) Init(versions []string) error {

	for _, version := range versions {
		var v *sv.Version
		var err error

		switch s.Strict {
		case true:
			v, err = sv.StrictNewVersion(version)
		case false:
			v, err = sv.NewVersion(version)
		}

		if err != nil {
			logrus.Debugf("Skipping %q because %s, skipping", version, err)
		} else {
			s.versions = append(s.versions, v)
		}
	}

	if len(s.versions) > 0 {
		return nil
	}

	return ErrNoValidSemVerFound
}

// Sort re-order a list of versions with the newest version first
func (s *Semver) Sort() {
	sort.Sort(sort.Reverse(sv.Collection(s.versions)))
}

// Search returns the version matching pattern from a sorted list.
func (s *Semver) Search(versions []string) error {
	// We need to be sure that at least one version exist
	if len(versions) == 0 {
		return ErrNoVersionsFound
	}

	err := s.Init(versions)
	if err != nil {
		logrus.Error(err)
		return err
	}

	s.Sort()

	if len(s.Constraint) == 0 {
		s.FoundVersion.ParsedVersion = s.versions[0].String()
		s.FoundVersion.OriginalVersion = s.versions[0].Original()
		return nil
	}

	c, err := sv.NewConstraint(s.Constraint)
	if err != nil {
		return err
	}

	for _, v := range s.versions {

		if c.Check(v) {
			s.FoundVersion.ParsedVersion = v.String()
			s.FoundVersion.OriginalVersion = v.Original()
			break
		}
	}
	if len(s.FoundVersion.ParsedVersion) == 0 {
		return ErrNoVersionFound
	}

	return nil
}
