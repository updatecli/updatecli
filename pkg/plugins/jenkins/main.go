package jenkins

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/olblak/updateCli/pkg/plugins/github"
	"github.com/olblak/updateCli/pkg/plugins/maven"
)

// Jenkins defines parameters needed to retrieve latest Jenkins version
// based on a specific release
type Jenkins struct {
	Release string        // Defines the release name like latest or weekly
	Version string        // Defines a specific release version
	Github  github.Github // Github Parameter used to retrieve a Jenkins changelog
}

const (
	// STABLE represents a stable release type
	STABLE string = "stable"
	// WEEKLY represents a weekly release type
	WEEKLY string = "weekly"
	// WRONG represents a bad release name
	WRONG string = "unknown"
)

// GetVersions fetch every jenkins version from the maven repository
func GetVersions() (latest string, versions []string, err error) {
	m := maven.Maven{
		URL:        "repo.jenkins-ci.org",
		Repository: "releases",
		GroupID:    "org.jenkins-ci.main",
		ArtifactID: "jenkins-war",
	}

	URL := fmt.Sprintf("https://%s/%s/%s/%s/maven-metadata.xml",
		m.URL,
		m.Repository,
		strings.ReplaceAll(m.GroupID, ".", "/"),
		m.ArtifactID)

	req, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		return "", nil, err
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return "", nil, err
	}

	data := maven.Metadata{}

	latest = data.Versioning.Latest
	versions = data.Versioning.Versions.Version

	xml.Unmarshal(body, &data)

	return data.Versioning.Latest, data.Versioning.Versions.Version, nil

}

// ReleaseType return the release type of a version
func ReleaseType(version string) (string, error) {
	components := strings.Split(version, ".")
	for _, component := range components {
		if _, err := strconv.Atoi(component); err != nil {
			return WRONG, fmt.Errorf("In version '%v', component '%v' is not a valid integer",
				version, component)
		}
	}

	if len(components) == 2 {
		return WEEKLY, nil
	} else if len(components) == 3 {
		return STABLE, nil
	}
	return WRONG, fmt.Errorf("Version %v contains %v component(s) which doesn't correspond to any valid release type", version, len(components))
}
