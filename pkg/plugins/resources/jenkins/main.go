package jenkins

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
	"github.com/updatecli/updatecli/pkg/plugins/utils/mavenmetadata"
)

// Spec defines a specification for a "jenkins" resource
// parsed from an updatecli manifest file
type Spec struct {
	Release string      // Defines the release name like latest or weekly
	Version string      // Defines a specific release version (condition only)
	Github  github.Spec // Github Parameter used to retrieve a Jenkins changelog
}

// Jenkins defines a resource of kind "githubrelease"
type Jenkins struct {
	spec             Spec
	mavenMetaHandler mavenmetadata.Handler
	foundVersion     string
}

const (
	// STABLE represents a stable release type
	STABLE string = "stable"
	// WEEKLY represents a weekly release type
	WEEKLY string = "weekly"
	// WRONG represents a bad release name
	WRONG string = "unknown"
	// URL of the default Jenkins Maven metadata file
	jenkinsDefaultMetaURL string = "https://repo.jenkins-ci.org/releases/org/jenkins-ci/main/jenkins-war/maven-metadata.xml"
)

// New returns a new valid GitHubRelease object.
func New(spec interface{}) (*Jenkins, error) {
	var newSpec Spec

	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return &Jenkins{}, err
	}

	if newSpec.Release == "" {
		newSpec.Release = STABLE
	}

	err = newSpec.Validate()
	if err != nil {
		return &Jenkins{}, err
	}

	return &Jenkins{
		spec:             newSpec,
		mavenMetaHandler: mavenmetadata.New(jenkinsDefaultMetaURL),
	}, nil
}

// Validate run some validation on the Jenkins struct
func (s Spec) Validate() (err error) {
	if len(s.Release) == 0 && len(s.Version) == 0 {
		logrus.Debugln("Jenkins release type not defined, default set to stable")
		s.Release = "stable"
	} else if len(s.Release) == 0 && len(s.Version) != 0 {
		s.Release, err = ReleaseType(s.Version)
		logrus.Debugf("Jenkins release type not defined, guessing based on Version %s", s.Version)
		if err != nil {
			return err
		}
	}

	if s.Release != WEEKLY &&
		s.Release != STABLE {
		return fmt.Errorf("wrong Jenkins release type '%s', accepted values ['%s','%s']",
			s.Release, WEEKLY, STABLE)

	}
	return nil
}

// GetVersions fetch every jenkins version from the maven repository
func (j *Jenkins) getVersions() (latest string, versions []string, err error) {
	latest, err = j.mavenMetaHandler.GetLatestVersion()
	if err != nil {
		return "", nil, err
	}

	versions, err = j.mavenMetaHandler.GetVersions()
	if err != nil {
		return "", nil, err
	}

	return latest, versions, nil

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
