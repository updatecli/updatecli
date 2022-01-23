package jenkins

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/github"
	"github.com/updatecli/updatecli/pkg/plugins/maven"
)

// Jenkins defines parameters needed to retrieve latest Jenkins version
// based on a specific release
type Jenkins struct {
	Release string      // Defines the release name like latest or weekly
	Version string      // Defines a specific release version
	Github  github.Spec // Github Parameter used to retrieve a Jenkins changelog
}

const (
	// STABLE represents a stable release type
	STABLE string = "stable"
	// WEEKLY represents a weekly release type
	WEEKLY string = "weekly"
	// WRONG represents a bad release name
	WRONG string = "unknown"
)

// Validate run some validation on the Jenkins struct
func (j *Jenkins) Validate() (err error) {
	if len(j.Release) == 0 && len(j.Version) == 0 {
		logrus.Debugln("Jenkins release type not defined, default set to stable")
		j.Release = "stable"
	} else if len(j.Release) == 0 && len(j.Version) != 0 {
		j.Release, err = ReleaseType(j.Version)
		logrus.Debugf("Jenkins release type not defined, guessing based on Version %s", j.Version)
		if err != nil {
			return err
		}
	}

	if j.Release != WEEKLY &&
		j.Release != STABLE {
		return fmt.Errorf("wrong Jenkins release type '%s', accepted values ['%s','%s']",
			j.Release, WEEKLY, STABLE)

	}
	return nil
}

// GetVersions fetch every jenkins version from the maven repository
func GetVersions() (latest string, versions []string, err error) {
	m, err := maven.New(maven.Spec{
		URL:        "repo.jenkins-ci.org",
		Repository: "releases",
		GroupID:    "org.jenkins-ci.main",
		ArtifactID: "jenkins-war",
	})
	if err != nil {
		return "", []string{}, err
	}

	req, err := http.NewRequest("GET", m.RepositoryURL, nil)

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

	err = xml.Unmarshal(body, &data)

	if err != nil {
		return "", nil, err
	}

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
