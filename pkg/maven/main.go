package maven

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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

// IsTagPublished test if a specific version exist on the maven repository
func (m *Maven) IsTagPublished() bool {
	URL := fmt.Sprintf("https://%s/%s/%s/%s/maven-metadata.xml",
		m.URL,
		m.Repository,
		strings.ReplaceAll(m.GroupID, ".", "/"),
		m.ArtifactID)

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		fmt.Println(err)
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Println(err)
	}

	data := Metadata{}

	xml.Unmarshal(body, &data)

	for _, version := range data.Versioning.Versions.Version {
		if version == m.Version {
			fmt.Printf("\u2714 Version %s is available on Maven Repository\n", m.Version)
			return true
		}

	}

	fmt.Printf("\u2717 Version %s is not available on Maven Repository\n", m.Version)
	return false
}

// GetVersion return the latest version
func (m *Maven) GetVersion() string {
	URL := fmt.Sprintf("https://%s/%s/%s/%s/maven-metadata.xml",
		m.URL,
		m.Repository,
		strings.ReplaceAll(m.GroupID, ".", "/"),
		m.ArtifactID)

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		fmt.Println(err)
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Println(err)
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		fmt.Println(err)
	}

	data := Metadata{}

	xml.Unmarshal(body, &data)

	if data.Versioning.Latest != "" {
		fmt.Printf("\u2714 Latest version is %s on Maven Repository\n", data.Versioning.Latest)
		return data.Versioning.Latest
	}

	fmt.Printf("\u2717 No latest version on Maven Repository\n")
	return ""
}
