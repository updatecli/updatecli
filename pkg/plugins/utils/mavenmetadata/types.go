package mavenmetadata

import (
	"encoding/xml"
)

// MetadataHandler must be implemented by any Maven metadata retriever
type Handler interface {
	GetMetadataURL() string
	GetLatestVersion() (string, error)
	GetVersions() ([]string, error)
}

// metadata hold maven repository Metadata
type metadata struct {
	Metadata   xml.Name `xml:"metadata"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
	Versioning version  `xml:"versioning"`
}

// version hold Version information
type version struct {
	Versioning xml.Name `xml:"versioning"`
	Latest     string   `xml:"latest"`
	Release    string   `xml:"release"`
	Versions   versions `xml:"versions"`
}

// versions contains the list of available version
type versions struct {
	ID      xml.Name `xml:"versions"`
	Version []string `xml:"version"`
}
