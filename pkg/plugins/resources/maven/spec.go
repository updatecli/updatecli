package maven

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// Spec defines a specification for a "maven" resource
// parsed from an updatecli manifest file
type Spec struct {
	// Deprecated, please specify the Maven url in the repository
	URL string `yaml:",omitempty"`
	// Specifies the maven repository url + name
	Repository string `yaml:",omitempty"`
	// Repositories specifies a list of Maven repository where to look for version. Order matter, version is retrieve from the first repository with the last one being Maven Central.
	Repositories []string `yaml:",omitempty"`
	// Specifies the maven artifact groupID
	GroupID string `yaml:",omitempty"`
	// Specifies the maven artifact artifactID
	ArtifactID string `yaml:",omitempty"`
	// Specifies the maven artifact version
	Version string `yaml:",omitempty"`
}

func (s Spec) Atomic() Spec {
	return Spec{
		ArtifactID:   s.ArtifactID,
		GroupID:      s.GroupID,
		Repository:   s.Repository,
		Repositories: s.Repositories,
		URL:          s.URL,
		Version:      s.Version,
	}
}

func (s *Spec) Sanitize() error {

	var errs []error
	var err error

	if len(s.URL) > 0 {
		logrus.Warningf("Parameter %q is deprecate, please prefix its content to parameter %q", "URL", "repository")
		s.Repository, err = joinURL([]string{s.URL, s.Repository})
		if err != nil {
			logrus.Errorln(err)
		}
	}

	if len(s.Repository) > 0 {
		sanitizedURL, err := joinURL([]string{s.Repository})
		if err != nil {
			errs = append(errs, err)
		} else {
			s.Repository = sanitizedURL
		}
	}

	for i := range s.Repositories {
		sanitizedURL, err := joinURL([]string{s.Repositories[i]})
		if err != nil {
			errs = append(errs, err)
			continue
		}
		s.Repositories[i] = sanitizedURL
	}

	if len(errs) > 0 {
		for i := range errs {
			logrus.Errorf("%s", errs[i])
		}
		return fmt.Errorf("failed sanitizing Maven spec")

	}

	return nil
}
