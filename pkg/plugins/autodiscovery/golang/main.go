package golang

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
	"github.com/updatecli/updatecli/pkg/plugins/utils/vulnerability"
)

/*
"golang" defines the specification for the Golang autodiscovery crawler.
It searches go.mod files and generates manifests to update the Go version and the Go modules.
*/
type Spec struct {
	// "rootdir" defines the directory where the crawler starts searching for go.mod files.
	//
	// default:
	//   the scm directory when "scmid" is set, otherwise the directory relative paths resolve from, by default the working directory.
	//
	// remark:
	//   * a relative path is resolved from the default directory.
	//   * an absolute path is used as is, instead of the scm directory.
	//
	RootDir string `yaml:",omitempty"`
	// "onlygoversion" restricts the autodiscovery to the Go version defined in go.mod.
	//
	// default:
	//   false
	//
	// remark:
	//   * it cannot be combined with "vulnerability".
	//
	OnlyGoVersion *bool `yaml:",omitempty"`
	// "onlygomodule" restricts the autodiscovery to the Go modules defined in go.mod.
	//
	// default:
	//   false
	//
	OnlyGoModule *bool `yaml:",omitempty"`
	// "ignore" defines rules to exclude matching Go modules or Go versions from the autodiscovery.
	//
	// remark:
	//   * a Go module or Go version is ignored when it matches at least one rule.
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// "only" defines rules to restrict the autodiscovery to matching Go modules or Go versions.
	//
	// remark:
	//   * a Go module or Go version is kept only when it matches at least one rule.
	//   * when every rule sets "goversion" without "modules", only the Go version is updated.
	//   * when every rule sets "modules" without "goversion", only the Go modules are updated.
	//
	Only MatchingRules `yaml:",omitempty"`
	// "versionfilter" defines the version filter used by the generated manifests.
	//
	// default:
	//   kind "semver" with pattern "*", any version greater than or equal to the current one.
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
	//   * a module using a pseudo version ignores the filter and is updated to the latest version.
	//   * it cannot be combined with "vulnerability".
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
	// default:
	//   empty, no age filtering.
	//
	// remark:
	//   * it accepts a duration string, such as "24h", "7d" or "1w".
	//   * it cannot be combined with "vulnerability".
	//
	Age age.Spec `yaml:",omitempty"`
	// "vulnerability" switches the autodiscovery to security updates, based on the OSV database (https://osv.dev).
	//
	// Each Go module is updated to the lowest version without known vulnerabilities,
	// and left untouched when it has none.
	//
	// remark:
	//   * only security updates are generated, routine updates require a separate manifest.
	//   * it cannot be combined with "age", "versionfilter" and "onlygoversion".
	//   * the Go version and indirect modules are not covered.
	//   * labels, such as "security", are set on the action used by the manifest.
	//
	// example:
	//   ```
	//   vulnerability:
	//     minseverity: high
	//     ignore:
	//       - GO-2025-3503
	//   ```
	//
	Vulnerability *vulnerability.Spec `yaml:",omitempty"`
}

// Golang holds all information needed to generate golang manifest.
type Golang struct {
	onlyGoModule  bool
	onlygoVersion bool
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for go.mod file
	rootDir string
	// actionID holds the actionID used by the newly generated manifest
	actionID string
	// scmID holds the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid object.
func New(spec interface{}, rootDir, scmID, actionID string) (Golang, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Golang{}, err
	}

	// Validate ignore rules
	if err := s.Ignore.Validate(); err != nil {
		return Golang{}, fmt.Errorf("invalid ignore spec: %w", err)
	}

	// Validate only rules
	if err := s.Only.Validate(); err != nil {
		return Golang{}, fmt.Errorf("invalid only spec: %w", err)
	}

	if err := s.validateVulnerability(); err != nil {
		return Golang{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		logrus.Debugln("no versioning filtering specified, fallback to semantic versioning")
		// By default, golang versioning uses semantic versioning
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	goVersionOnly := false
	if s.OnlyGoVersion != nil {
		goVersionOnly = *s.OnlyGoVersion
	}

	goModuleOnly := false
	if s.OnlyGoModule != nil {
		goModuleOnly = *s.OnlyGoModule
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
		return Golang{}, err
	}

	g := Golang{
		spec:          s,
		rootDir:       dir,
		scmID:         scmID,
		onlyGoModule:  goModuleOnly,
		onlygoVersion: goVersionOnly,
		versionFilter: newFilter,
		actionID:      actionID,
	}

	return g, nil

}

// validateVulnerability checks the vulnerability spec and the settings it can't be combined with.
func (s Spec) validateVulnerability() error {
	if s.Vulnerability == nil {
		return nil
	}

	if err := s.Vulnerability.Validate(); err != nil {
		return fmt.Errorf("invalid vulnerability spec: %w", err)
	}

	if s.Age.Minimum != "" || s.Age.Maximum != "" {
		return errors.New("age can't be combined with vulnerability, use a separate manifest for routine updates")
	}

	if !s.VersionFilter.IsZero() {
		return errors.New("versionfilter can't be combined with vulnerability, the version is the lowest one without known vulnerabilities")
	}

	if (s.OnlyGoVersion != nil && *s.OnlyGoVersion) || s.Only.isGoVersionOnly() {
		return errors.New("onlygoversion can't be combined with vulnerability, Go version security updates are not supported")
	}

	return nil
}

func (n Golang) DiscoverManifests() ([][]byte, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Golang"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Golang")+1))

	manifests, err := n.discoverDependencyManifests()

	if err != nil {
		return nil, err
	}

	return manifests, nil
}
