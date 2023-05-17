package maven

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/mavenmetadata"
)

var (
	// ErrWrongSpec is returned when the Spec has wrong content
	ErrWrongSpec error = errors.New("wrong spec content")

	MavenCentralRepository string = "https://repo1.maven.org/maven2/"
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

// Maven defines a resource of kind "maven"
type Maven struct {
	spec             Spec
	metadataHandlers []mavenmetadata.Handler
}

// New returns a reference to a newly initialized Maven object from a Spec
// or an error if the provided Spec triggers a validation error.
func New(spec interface{}) (*Maven, error) {
	newSpec := Spec{}
	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return &Maven{}, nil
	}

	err = newSpec.Sanitize()

	if err != nil {
		return &Maven{}, nil
	}

	newResource := &Maven{
		spec: newSpec,
	}

	if len(newSpec.Repository) > 0 {

		u, err := url.Parse(newSpec.Repository)
		if err != nil {
			return &Maven{}, err
		}

		u.Path = path.Join(
			u.Path,
			strings.ReplaceAll(newSpec.GroupID, ".", "/"),
			newSpec.ArtifactID,
			"maven-metadata.xml")

		newResource.metadataHandlers = append(
			newResource.metadataHandlers,
			mavenmetadata.New(u.String()))

		return newResource, nil
	}

	for i := range newSpec.Repositories {
		u, err := url.Parse(newSpec.Repositories[i])
		if err != nil {
			return &Maven{}, err
		}

		u.Path = path.Join(
			u.Path,
			strings.ReplaceAll(newSpec.GroupID, ".", "/"),
			newSpec.ArtifactID,
			"maven-metadata.xml")

		newResource.metadataHandlers = append(
			newResource.metadataHandlers,
			mavenmetadata.New(u.String()))
	}

	mavenCentralNotFound, err := isRepositoriesContainsMavenCentral(newSpec.Repositories)

	if err != nil {
		return &Maven{}, err
	}

	if !mavenCentralNotFound {
		u, err := url.Parse(MavenCentralRepository)
		if err != nil {
			return &Maven{}, err
		}

		u.Path = path.Join(u.Path, strings.ReplaceAll(newSpec.GroupID, ".", "/"), newSpec.ArtifactID, "maven-metadata.xml")

		newResource.metadataHandlers = append(
			newResource.metadataHandlers,
			mavenmetadata.New(u.String()))
	}

	return newResource, nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (m *Maven) Changelog() string {
	return ""
}

func (m Maven) Validate() error {
	errs := []error{}

	if len(m.spec.Repository) > 0 && len(m.spec.Repositories) > 0 {
		errs = append(errs, fmt.Errorf("parameter %q and %q are mutually exclusive",
			"repository",
			"repositories"))
	}

	if len(errs) > 0 {
		for _, e := range errs {
			logrus.Errorf(e.Error())
		}
		return ErrWrongSpec
	}
	return nil

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
