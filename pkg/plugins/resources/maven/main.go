package maven

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/updatecli/updatecli/pkg/plugins/utils/mavenmetadata"
)

// Spec defines a specification for a "maven" resource
// parsed from an updatecli manifest file
type Spec struct {
	URL        string
	Repository string
	GroupID    string
	ArtifactID string
	Version    string
}

// Maven defines a resource of kind "maven"
type Maven struct {
	spec            Spec
	metadataHandler mavenmetadata.Handler
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
		metadataHandler: mavenmetadata.New(
			fmt.Sprintf("https://%s/%s/%s/%s/maven-metadata.xml",
				newSpec.URL,
				newSpec.Repository,
				strings.ReplaceAll(newSpec.GroupID, ".", "/"),
				newSpec.ArtifactID),
		),
	}

	return newResource, nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (m *Maven) Changelog() string {
	return ""
}
