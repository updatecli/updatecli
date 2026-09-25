package maven

import (
	"fmt"
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"maven" defines the specification for the Maven autodiscovery crawler.
It searches pom.xml files and generates manifests to update their dependencies, managed dependencies and parent pom.
*/
type Spec struct {
	// "rootdir" defines the directory where the crawler starts searching for pom.xml files.
	//
	// default:
	//   the scm directory when "scmid" is set, otherwise the directory relative paths resolve from, by default the working directory.
	//
	// remark:
	//   * a relative path is resolved from the default directory.
	//   * an absolute path is used as is, instead of the scm directory.
	//
	RootDir string `yaml:",omitempty"`
	// "ignore" defines rules to exclude matching Maven dependencies from the autodiscovery.
	//
	// remark:
	//   * a Maven dependency is ignored when it matches at least one rule.
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// "only" defines rules to restrict the autodiscovery to matching Maven dependencies.
	//
	// remark:
	//   * a Maven dependency is kept only when it matches at least one rule.
	//
	Only MatchingRules `yaml:",omitempty"`
	// "versionfilter" defines the version filter used by the generated manifests.
	//
	// default:
	//   kind "latest", the latest version.
	//
	// remark:
	//   * with kind "semver", "pattern" accepts:
	//     * "prerelease": the latest prerelease of the current version.
	//     * "patch": patch updates only.
	//     * "minor": patch and minor updates.
	//     * "minoronly": minor updates only.
	//     * "major": patch, minor and major updates.
	//     * "majoronly": major updates only.
	//     * a version constraint, such as ">= 1.0.0".
	//   * with kind "regex", "pattern" accepts a regular expression.
	//   * more examples at https://www.updatecli.io/docs/core/versionfilter/
	//
	// example:
	//   ```
	//   versionfilter:
	//     kind: semver
	//     pattern: minor
	//   ```
	//
	VersionFilter version.Filter `yaml:",omitempty"`
}

// Maven hold all information needed to generate helm manifest.
type Maven struct {
	// actionID holds the actionID used by the newly generated manifest
	actionID string
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for Helm Chart
	rootDir string
	// scmID holds the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid Helm object.
func New(spec interface{}, rootDir, scmID, actionID string) (Maven, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Maven{}, err
	}

	// Validate ignore rules
	if err := s.Ignore.Validate(); err != nil {
		return Maven{}, fmt.Errorf("invalid ignore spec: %w", err)
	}

	// Validate only rules
	if err := s.Only.Validate(); err != nil {
		return Maven{}, fmt.Errorf("invalid only spec: %w", err)
	}

	dir := rootDir
	if path.IsAbs(s.RootDir) {
		if scmID != "" {
			logrus.Warningf("rootdir %q is an absolute path, scmID %q will be ignored", s.RootDir, scmID)
		}
		dir = s.RootDir
	}

	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return Maven{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		newFilter.Kind = "latest"
		newFilter.Pattern = "latest"
	}

	return Maven{
		actionID:      actionID,
		spec:          s,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
	}, nil

}

func (m Maven) DiscoverManifests() ([][]byte, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Maven"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Maven")+1))

	manifests, err := m.discoverDependencyManifests("dependency")

	if err != nil {
		return nil, err
	}

	dependencyManagementManifests, err := m.discoverDependencyManifests("dependencyManagement")

	if err != nil {
		return nil, err
	}

	manifests = append(manifests, dependencyManagementManifests...)

	parentPomdependencyManifests, err := m.discoverParentPomDependencyManifests()

	if err != nil {
		return nil, err
	}

	manifests = append(manifests, parentPomdependencyManifests...)

	return manifests, nil
}
