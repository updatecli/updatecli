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

	MavenCentralRepository string = "repo1.maven.org/maven2/"
)

// Spec defines a specification for a "maven" resource
// parsed from an updatecli manifest file
type Spec struct {
	// Specifies the maven repository URL
	URL string `yaml:",omitempty"`
	// Specifies the maven repository name
	Repository string `yaml:",omitempty"`
	// Repositories specifies a list of Maven repository
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

	newResource := &Maven{
		spec: newSpec,
	}

	if len(newSpec.Repository) > 0 {

		URL := strings.TrimPrefix(newSpec.URL, "https://")

		URL = strings.TrimPrefix(URL, "http://")

		u, err := url.Parse("https://" + URL)
		if err != nil {
			return &Maven{}, err
		}

		u.Path = path.Join(
			u.Path,
			newSpec.Repository,
			strings.ReplaceAll(newSpec.GroupID, ".", "/"),
			newSpec.ArtifactID,
			"maven-metadata.xml")

		newResource.metadataHandlers = append(
			newResource.metadataHandlers,
			mavenmetadata.New(u.String()))

		return newResource, nil
	}

	for _, repository := range newSpec.Repositories {
		URL := strings.TrimPrefix(repository, "https://")
		URL = strings.TrimPrefix(URL, "http://")

		u, err := url.Parse("https://" + URL)
		if err != nil {
			return &Maven{}, err
		}

		u.Path = path.Join(u.Path, strings.ReplaceAll(newSpec.GroupID, ".", "/"), newSpec.ArtifactID, "maven-metadata.xml")

		newResource.metadataHandlers = append(
			newResource.metadataHandlers,
			mavenmetadata.New(u.String()))
	}

	mavenCentrallNotFound, err := isRepositoriesContainsMavenCentral(newSpec.Repositories)

	if err != nil {
		return &Maven{}, err
	}

	if !mavenCentrallNotFound {
		u, err := url.Parse("https://" + MavenCentralRepository)
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
		errs = append(errs, fmt.Errorf("parameter %q and %q are mutually exclusif",
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
