package maven

import (
	"encoding/xml"
)

// Maven hold maven repository information
type Maven struct {
	URL        string
	Repository string
	GroupID    string
	ArtifactID string
	Version    string
}

// Metadata hold maven repository metadata
type Metadata struct {
	Metadata   xml.Name `xml:"metadata"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
	Versioning Version  `xml:"versioning"`
}

// Version hold version information
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
