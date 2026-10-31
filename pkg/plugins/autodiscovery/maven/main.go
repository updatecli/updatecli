package maven

import (
	"errors"
	"fmt"
	"path"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
	"github.com/updatecli/updatecli/pkg/plugins/utils/vulnerability"
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
	// "age" defines the minimum or maximum age of a release to be considered valid.
	//
	// remark:
	//   * it accepts a duration string, such as "24h", "7d", "3w" or "1y".
	//   * it is propagated to the generated maven sources.
	//   * a release's age is the Last-Modified date the repository reports for its pom file.
	//     On a repository reporting none, the generated sources are skipped.
	//   * when the most recent release is too recent, an older one is suggested, which can be lower
	//     than the version in pom.xml. Combine it with a "versionfilter" to bound the versions.
	//   * security fixes don't wait for it, see "vulnerability".
	//
	// example:
	//   ```
	//   age:
	//     minimum: 7d
	//   ```
	//
	Age age.Spec `yaml:",omitempty"`
	// "vulnerability" switches the autodiscovery to security updates, based on the OSV database (https://osv.dev).
	//
	// A Maven artifact is only updated when its current version has known vulnerabilities,
	// to the lowest version without any.
	//
	// remark:
	//   * only security updates are generated, routine updates require a separate manifest.
	//   * it cannot be combined with:
	//     * "age", as a fixed version is suggested as soon as it is known.
	//     * "versionfilter", as the version is the lowest one without known vulnerabilities.
	//   * dependencies, managed dependencies and the parent pom are covered.
	//   * the current version is the version written in pom.xml. When it is a property, such as
	//     "${log4j.version}", defined in the same pom.xml, the property is updated instead.
	//     Artifacts relying on a property defined elsewhere, or on a version range, such as "[1.0,2.0)", are ignored.
	//   * several artifacts sharing a property each get their own manifest updating that property.
	//     Once one manifest updates it, the others are skipped, and the next run checks the new version.
	//   * BOM imports are ignored, as OSV knows about the artifacts they manage, not the BOM itself.
	//   * labels, such as "security", are set on the action used by the manifest.
	//
	// example:
	//   ```
	//   vulnerability:
	//     minseverity: high
	//     ignore:
	//       - CVE-2021-44228
	//   ```
	//
	Vulnerability *vulnerability.Spec `yaml:",omitempty"`
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

	if err := s.validateVulnerability(); err != nil {
		return Maven{}, err
	}

	if err := s.Age.Validate(); err != nil {
		return Maven{}, fmt.Errorf("wrong age spec: %w", err)
	}

	dir := rootDir
	if path.IsAbs(s.RootDir) {
		if scmID != "" {
			log.Warningf("rootdir %q is an absolute path, scmID %q will be ignored", s.RootDir, scmID)
		}
		dir = s.RootDir
	}

	if len(dir) == 0 {
		log.Errorln("no working directory defined")
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

// validateVulnerability checks the vulnerability spec and the settings it can't be combined with.
func (s Spec) validateVulnerability() error {
	if s.Vulnerability == nil {
		return nil
	}

	if err := s.Vulnerability.Validate(); err != nil {
		return fmt.Errorf("invalid vulnerability spec: %w", err)
	}

	if !s.Age.IsZero() {
		return errors.New("age can't be combined with vulnerability, use a separate manifest for routine updates")
	}

	if !s.VersionFilter.IsZero() {
		return errors.New("versionfilter can't be combined with vulnerability, the version is the lowest one without known vulnerabilities")
	}

	return nil
}

func (m Maven) DiscoverManifests() ([][]byte, error) {

	log.Heading("Maven")

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
