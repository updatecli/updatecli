package maven

import (
	"fmt"
	"strings"

	"github.com/updatecli/updatecli/pkg/plugins/mavenmetadata"
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
func New(spec Spec) (*Maven, error) {
	newResource := &Maven{
		spec: spec,
		metadataHandler: mavenmetadata.New(
			fmt.Sprintf("https://%s/%s/%s/%s/maven-metadata.xml",
				spec.URL,
				spec.Repository,
				strings.ReplaceAll(spec.GroupID, ".", "/"),
				spec.ArtifactID),
		),
	}

	return newResource, nil
}
