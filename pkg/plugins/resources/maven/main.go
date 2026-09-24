package maven

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/mavenmetadata"
	"github.com/updatecli/updatecli/pkg/plugins/utils/redact"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

var (
	// ErrWrongSpec is returned when the Spec has wrong content
	ErrWrongSpec error = errors.New("wrong spec content")

	MavenCentralRepository string = "https://repo1.maven.org/maven2/"
)

/*
"maven" defines the specification for retrieving Maven artifact versions.
It can be used as a "source" or a "condition".
*/
type Spec struct {
	// "url" defines the Maven repository base url.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// deprecated:
	//   * set the full url in "repository" instead. "url" is prefixed to "repository".
	//
	URL string `yaml:",omitempty"`
	// "repository" defines the Maven repository url, including the repository name.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// remark:
	//   * "repository" and "repositories" are mutually exclusive. When both are set, "repositories" is ignored.
	//   * "https://" is added when the url has no scheme.
	//   * Maven Central is not queried when "repository" is set.
	//
	// example:
	//   * repository: https://repo.jenkins-ci.org/releases
	//
	Repository string `yaml:",omitempty"`
	// "repositories" defines the list of Maven repositories where to look for versions.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// default:
	//   Maven Central
	//
	// remark:
	//   * "repository" and "repositories" are mutually exclusive.
	//   * order matters: the version is retrieved from the first repository that returns one.
	//   * Maven Central is added as the last repository unless the list already holds it.
	//   * "https://" is added when a url has no scheme.
	//
	// example:
	//   * repositories:
	//     - https://repo.jenkins-ci.org/releases
	//     - https://repo1.maven.org/maven2
	//
	Repositories []string `yaml:",omitempty"`
	// "groupid" defines the Maven artifact groupId.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// example:
	//   * groupid: org.jenkins-ci.main
	//
	GroupID string `yaml:",omitempty"`
	// "artifactid" defines the Maven artifact artifactId.
	//
	// compatible:
	//   * source
	//   * condition
	//
	// example:
	//   * artifactid: jenkins-war
	//
	ArtifactID string `yaml:",omitempty"`
	// "version" defines the Maven artifact version to check.
	//
	// compatible:
	//   * condition
	//
	// default:
	//   the output of the associated source.
	//
	Version string `yaml:",omitempty"`
	// "versionfilter" defines the version pattern and its kind, such as "regex", "semver" or "latest".
	//
	// compatible:
	//   * source
	//
	VersionFilter version.Filter `yaml:",omitempty"`
}

// Maven defines a resource of kind "maven"
type Maven struct {
	spec             Spec
	metadataHandlers []mavenmetadata.Handler
}

// New returns a reference to a newly initialized Maven object from a Spec
// or an error if the provided Spec triggers a validation error.
func New(spec interface{}) (*Maven, error) {
	newSpec := Spec{}
	err := mapstructure.Decode(spec, &newSpec)
	if err != nil {
		return &Maven{}, err
	}

	err = newSpec.Sanitize()

	if err != nil {
		return &Maven{}, err
	}

	newResource := &Maven{
		spec: newSpec,
	}

	if len(newSpec.Repository) > 0 {

		u, err := url.Parse(newSpec.Repository)
		if err != nil {
			return &Maven{}, err
		}

		u.Path = path.Join(
			u.Path,
			strings.ReplaceAll(newSpec.GroupID, ".", "/"),
			newSpec.ArtifactID,
			"maven-metadata.xml")

		newResource.metadataHandlers = append(
			newResource.metadataHandlers,
			mavenmetadata.New(u.String(), newSpec.VersionFilter))

		return newResource, nil
	}

	for i := range newSpec.Repositories {
		u, err := url.Parse(newSpec.Repositories[i])
		if err != nil {
			return &Maven{}, err
		}

		u.Path = path.Join(
			u.Path,
			strings.ReplaceAll(newSpec.GroupID, ".", "/"),
			newSpec.ArtifactID,
			"maven-metadata.xml")

		newResource.metadataHandlers = append(
			newResource.metadataHandlers,
			mavenmetadata.New(u.String(), newSpec.VersionFilter))
	}

	mavenCentralNotFound, err := isRepositoriesContainsMavenCentral(newSpec.Repositories)

	if err != nil {
		return &Maven{}, err
	}

	if !mavenCentralNotFound {
		u, err := url.Parse(MavenCentralRepository)
		if err != nil {
			return &Maven{}, err
		}

		u.Path = path.Join(u.Path, strings.ReplaceAll(newSpec.GroupID, ".", "/"), newSpec.ArtifactID, "maven-metadata.xml")

		newResource.metadataHandlers = append(
			newResource.metadataHandlers,
			mavenmetadata.New(u.String(), newSpec.VersionFilter))
	}

	return newResource, nil
}

// Changelog returns the changelog for this resource, or an empty string if not supported
func (m *Maven) Changelog(from, to string) *result.Changelogs {
	return nil
}

func (m Maven) Validate() error {
	errs := []error{}

	if len(m.spec.Repository) > 0 && len(m.spec.Repositories) > 0 {
		errs = append(errs, fmt.Errorf("parameter %q and %q are mutually exclusive",
			"repository",
			"repositories"))
	}

	if len(errs) > 0 {
		for _, e := range errs {
			logrus.Errorln(e.Error())
		}
		return ErrWrongSpec
	}
	return nil

}

func (s *Spec) Sanitize() error {

	var errs []error
	var err error

	if len(s.URL) > 0 {
		logrus.Warningf("Parameter %q is deprecate, please prefix its content to parameter %q", "URL", "repository")
		s.Repository, err = joinURL([]string{s.URL, s.Repository})
		if err != nil {
			logrus.Errorln(err)
		}
	}

	if len(s.Repository) > 0 {
		sanitizedURL, err := joinURL([]string{s.Repository})
		if err != nil {
			errs = append(errs, err)
		} else {
			s.Repository = sanitizedURL
		}
	}

	for i := range s.Repositories {
		sanitizedURL, err := joinURL([]string{s.Repositories[i]})
		if err != nil {
			errs = append(errs, err)
			continue
		}
		s.Repositories[i] = sanitizedURL
	}

	if len(errs) > 0 {
		for i := range errs {
			logrus.Errorf("%s", errs[i])
		}
		return fmt.Errorf("failed sanitizing Maven spec")

	}

	return nil
}

// ReportConfig returns a new configuration with only the necessary configuration fields
// to identify the resource without any sensitive information
// and context specific data.
func (m *Maven) ReportConfig() interface{} {

	repositories := make([]string, len(m.spec.Repositories))
	for i := range m.spec.Repositories {
		repositories[i] = redact.URL(m.spec.Repositories[i])
	}

	return Spec{
		GroupID:      m.spec.GroupID,
		ArtifactID:   m.spec.ArtifactID,
		Version:      m.spec.Version,
		Repository:   redact.URL(m.spec.Repository),
		Repositories: repositories,
		URL:          redact.URL(m.spec.URL),
	}
}
