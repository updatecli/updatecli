package npm

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
"npm" defines the specification for the npm autodiscovery crawler.
It searches package.json files and generates manifests to update their npm packages.
*/
type Spec struct {
	// "rootdir" defines the directory where the crawler starts searching for package.json files.
	//
	// default:
	//   the scm directory when "scmid" is set, otherwise the directory relative paths resolve from, by default the working directory.
	//
	// remark:
	//   * a relative path is resolved from the default directory.
	//   * an absolute path is used as is, instead of the scm directory.
	//
	RootDir string `yaml:",omitempty"`
	// "ignore" defines rules to exclude matching npm packages from the autodiscovery.
	//
	// remark:
	//   * a npm package is ignored when it matches at least one rule.
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// "only" defines rules to restrict the autodiscovery to matching npm packages.
	//
	// remark:
	//   * a npm package is kept only when it matches at least one rule.
	//
	Only MatchingRules `yaml:",omitempty"`
	// "versionfilter" defines the version filter used by the generated manifests.
	//
	// default:
	//   * an exact version: kind "semver" with pattern ">=<current version>".
	//   * a version constraint: kind "semver" with the constraint as pattern, or "*" when "ignoreversionconstraints" is true.
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
	//   * it is ignored for a package declared with a version constraint, unless "ignoreversionconstraints" is true.
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
	// "ignoreversionconstraints" defines whether the version constraints set in package.json are ignored.
	//
	// When true, a package declared with a version constraint is updated to the latest version
	// accepted by "versionfilter", instead of the latest version accepted by the constraint.
	//
	// default:
	//   false
	//
	// remark:
	//   * when true and "versionfilter" is set, Updatecli converts the constraint to a version
	//     so "versionfilter" can select the next patch, minor or major version. A complex constraint,
	//     such as ">=1.0.0 <2.0.0", is converted to the first version it holds, 1.0.0 in this example.
	//   * it cannot be combined with "vulnerability".
	//
	IgnoreVersionConstraints *bool `yaml:",omitempty"`
	// "npmrcpath" defines the path of the .npmrc file used by every discovered package.
	//
	// remark:
	//   * it is propagated to the generated npm resources.
	//
	NpmrcPath string `yaml:"npmrcpath,omitempty"`
	// "url" defines the npm registry url.
	//
	// default:
	//   https://registry.npmjs.org/
	//
	// remark:
	//   * it is propagated to the generated npm sources.
	//
	URL string `yaml:",omitempty"`
	// "registrytoken" defines the token used to authenticate with the registry.
	//
	// remark:
	//   * it is propagated to the generated npm sources.
	//
	RegistryToken string `yaml:",omitempty"`
	// "age" defines the minimum or maximum age of a release to be considered valid.
	//
	// default:
	//   ```
	//   age:
	//     minimum: 3d
	//   ```
	//
	// remark:
	//   * it accepts a duration string, such as "24h", "7d", "3w" or "1y".
	//   * it is propagated to the generated npm sources.
	//   * the default keeps Updatecli from suggesting a package version published less than three days ago.
	//   * an empty "age: {}" disables the default.
	//   * it cannot be combined with "vulnerability".
	//
	Age *age.Spec `yaml:",omitempty"`
	// "vulnerability" switches the autodiscovery to security updates, based on the OSV database (https://osv.dev).
	//
	// A package is only updated when its current version has known vulnerabilities,
	// to the lowest version without any.
	//
	// remark:
	//   * only security updates are generated, routine updates require a separate manifest.
	//   * it cannot be combined with "age", "versionfilter" and "ignoreversionconstraints".
	//   * the default minimum release age does not apply, a fixed version is suggested as soon as it is known.
	//   * the current version is the exact version from package.json or, for a version constraint,
	//     the version resolved in the package-lock.json, pnpm-lock.yaml or yarn.lock next to it,
	//     or at the root of its workspace. Packages without an identifiable current version are ignored.
	//   * labels, such as "security", are set on the action used by the manifest.
	//
	// example:
	//   ```
	//   vulnerability:
	//     minseverity: high
	//     ignore:
	//       - GHSA-jr5f-v2jv-69x6
	//   ```
	//
	Vulnerability *vulnerability.Spec `yaml:",omitempty"`
}

// Npm holds all information needed to generate npm manifest.
type Npm struct {
	// actionID holds the actionID used by the newly generated manifest
	actionID string
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for NPM
	rootDir string
	// scmID holds the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	// ignoreVersionConstraint indicates whether to respect version constraints defined in package.json or not.
	ignoreVersionConstraint bool
	// npmrcPath holds the path to the .npmrc file to propagate to generated manifests
	npmrcPath string
	// url holds the registry URL to propagate to generated manifests
	url string
	// registryToken holds the registry token to propagate to generated manifests
	registryToken string
	// releaseAge holds the release age filter to propagate to generated manifests
	releaseAge age.Spec
}

// New return a new valid object.
func New(spec interface{}, rootDir, scmID, actionID string) (Npm, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Npm{}, err
	}

	// Validate ignore rules
	if err := s.Ignore.Validate(); err != nil {
		return Npm{}, fmt.Errorf("invalid ignore spec: %w", err)
	}

	// Validate only rules
	if err := s.Only.Validate(); err != nil {
		return Npm{}, fmt.Errorf("invalid only spec: %w", err)
	}

	if err := s.validateVulnerability(); err != nil {
		return Npm{}, err
	}

	// By default, wait for a release to be old enough before suggesting it
	releaseAge := age.Spec{Minimum: defaultMinimumReleaseAge}
	switch {
	case s.Age != nil:
		releaseAge = *s.Age
	case s.Vulnerability != nil:
		// A fixed version is suggested as soon as it is known
		releaseAge = age.Spec{}
	}

	// Validate the release age filter
	if err := releaseAge.Validate(); err != nil {
		return Npm{}, fmt.Errorf("wrong age spec %v", err)
	}

	// By default we want to suggest the latest version available in the registry
	ignoreVersionConstraint := false
	if s.IgnoreVersionConstraints != nil {
		ignoreVersionConstraint = *s.IgnoreVersionConstraints
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
		return Npm{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, NPM versioning uses semantic versioning
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	return Npm{
		actionID:                actionID,
		spec:                    s,
		rootDir:                 dir,
		scmID:                   scmID,
		versionFilter:           newFilter,
		ignoreVersionConstraint: ignoreVersionConstraint,
		npmrcPath:               s.NpmrcPath,
		url:                     s.URL,
		registryToken:           s.RegistryToken,
		releaseAge:              releaseAge,
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

	if s.Age != nil && !s.Age.IsZero() {
		return errors.New("age can't be combined with vulnerability, use a separate manifest for routine updates")
	}

	if !s.VersionFilter.IsZero() {
		return errors.New("versionfilter can't be combined with vulnerability, the version is the lowest one without known vulnerabilities")
	}

	if s.IgnoreVersionConstraints != nil {
		return errors.New("ignoreversionconstraints can't be combined with vulnerability, the version is the lowest one without known vulnerabilities")
	}

	return nil
}

func (n Npm) DiscoverManifests() ([][]byte, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("NPM"))
	logrus.Infof("%s\n", strings.Repeat("=", len("NPM")+1))

	manifests, err := n.discoverDependencyManifests()

	if err != nil {
		return nil, err
	}

	return manifests, nil
}
