package maven

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/olblak/updateCli/pkg/scm"
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

// Condition tests if a specific version exist on the maven repository
func (m *Maven) Condition(source string) (bool, error) {

	if m.Version != "" {
		fmt.Printf("Version %v, already defined from configuration file\n", m.Version)
	} else {
		m.Version = source
	}
	URL := fmt.Sprintf("https://%s/%s/%s/%s/maven-metadata.xml",
		m.URL,
		m.Repository,
		strings.ReplaceAll(m.GroupID, ".", "/"),
		m.ArtifactID)

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		return false, err
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return false, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return false, err
	}

	data := Metadata{}

	xml.Unmarshal(body, &data)

	for _, version := range data.Versioning.Versions.Version {
		if version == m.Version {
			fmt.Printf("\u2714 Version %s is available on Maven Repository\n", m.Version)
			return true, nil
		}

	}

	fmt.Printf("\u2717 Version %s is not available on Maven Repository\n", m.Version)
	return false, nil
}

// ConditionFromSCM returns an error because it's not supported
func (m *Maven) ConditionFromSCM(source string, scm scm.Scm) (bool, error) {
	return false, fmt.Errorf("SCM configuration is not supported for maven condition, aborting")
}

// Source return the latest version
func (m *Maven) Source() (string, error) {
	URL := fmt.Sprintf("https://%s/%s/%s/%s/maven-metadata.xml",
		m.URL,
		m.Repository,
		strings.ReplaceAll(m.GroupID, ".", "/"),
		m.ArtifactID)

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		return "", err
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", err
	}

	data := Metadata{}

	xml.Unmarshal(body, &data)

	if data.Versioning.Latest != "" {
		fmt.Printf("\u2714 Latest version is %s on Maven Repository\n", data.Versioning.Latest)
		return data.Versioning.Latest, nil
	}

	fmt.Printf("\u2717 No latest version on Maven Repository\n")
	return "", nil
}
