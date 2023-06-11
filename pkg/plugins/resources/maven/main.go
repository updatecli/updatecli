package maven

import (
	"errors"
	"net/url"
	"path"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/plugins/utils/mavenmetadata"
)

var (
	// ErrWrongSpec is returned when the Spec has wrong content
	ErrWrongSpec error = errors.New("wrong spec content")

	MavenCentralRepository string = "https://repo1.maven.org/maven2/"
)

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
			mavenmetadata.New(u.String(), newSpec.VersionFilter))

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
			mavenmetadata.New(u.String(), newSpec.VersionFilter))
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
			mavenmetadata.New(u.String(), newSpec.VersionFilter))
	}

	return newResource, nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (m *Maven) Changelog() string {
	return ""
}
