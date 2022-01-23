package maven

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/httpclient"
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
	spec          Spec
	webClient     httpclient.HTTPClient
	RepositoryURL string
}

// New returns a reference to a newly initialized Maven object from a Spec
// or an error if the provided Spec triggers a validation error.
func New(spec Spec) (*Maven, error) {
	newResource := &Maven{
		spec:      spec,
		webClient: http.DefaultClient,
		RepositoryURL: fmt.Sprintf("https://%s/%s/%s/%s/maven-metadata.xml",
			spec.URL,
			spec.Repository,
			strings.ReplaceAll(spec.GroupID, ".", "/"),
			spec.ArtifactID),
	}

	return newResource, nil
}

// Metadata hold maven repository Metadata
type Metadata struct {
	Metadata   xml.Name `xml:"metadata"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
	Versioning Version  `xml:"versioning"`
}

// Version hold Version information
type Version struct {
	Versioning xml.Name `xml:"versioning"`
	Latest     string   `xml:"latest"`
	Release    string   `xml:"release"`
	Versions   Versions `xml:"versions"`
}

// Versions contains the list of available version
type Versions struct {
	ID      xml.Name `xml:"versions"`
	Version []string `xml:"version"`
}
