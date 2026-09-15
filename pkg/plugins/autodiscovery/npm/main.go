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

// Spec defines the parameters which can be provided to the NPM builder.
type Spec struct {
	// RootDir defines the root directory used to recursively search for npm packages.json
	RootDir string `yaml:",omitempty"`
	// Ignore allows to specify rule to ignore autodiscovery a specific NPM based on a rule
	Ignore MatchingRules `yaml:",omitempty"`
	// Only allows to specify rule to only autodiscover manifest for a specific NPM based on a rule
	Only MatchingRules `yaml:",omitempty"`
	//  `versionfilter` provides parameters to specify the version pattern used when generating manifest.
	//
	//  kind - semver
	//    versionfilter of kind `semver` uses semantic versioning as version filtering
	//    pattern accepts one of:
	//      `prerelease` - Updatecli tries to identify the latest prerelease whatever it means
	//      `patch` - Updatecli only handles patch version update
	//      `minor` - Updatecli handles patch AND minor version update
	//      `minoronly` - Updatecli handles minor version only
	//      `major` - Updatecli handles patch, minor, AND major version update
	//      `majoronly` - Updatecli only handles major version update
	//      `a version constraint` such as `>= 1.0.0`
	//
	//  kind - regex
	//    versionfilter of kind `regex` uses regular expression as version filtering
	//    pattern accepts a valid regular expression
	//
	//  example:
	//  ```
	//    versionfilter:
	//      kind: semver
	//      pattern: minor
	//  ```
	//
	//  and its type like regex, semver, or just latest.
	//
	//  More examples can be found at https://www.updatecli.io/docs/core/versionfilter/
	VersionFilter version.Filter `yaml:",omitempty"`
	// IgnoreVersionConstraints indicates whether to respect version constraints defined in package.json or not.
	// When set to true, Updatecli will ignore version constraints and update to the latest version available
	// in the registry according to the specified version filter.
	// Default is false.
	//
	// Remark:
	//  * If set to false, Updatecli will try to convert version constrains to valid semantic version
	//    so we can use versionFilter to retrieve the last Major/Minor/Patch version but in case of complex version constraints, such as `>=1.0.0 <2.0.0`,
	//    Updatecli will convert it to the first version it detects such as 1.0.0 in our example
	IgnoreVersionConstraints *bool `yaml:",omitempty"`
	// NpmrcPath defines the path to the .npmrc file to use for all discovered packages.
	// This will be propagated to all generated npm resource specs.
	NpmrcPath string `yaml:"npmrcpath,omitempty"`
	// URL defines the registry url (defaults to `https://registry.npmjs.org/`).
	// This will be propagated to all generated npm resource specs.
	URL string `yaml:",omitempty"`
	// RegistryToken defines the token to use when connecting to the registry.
	// This will be propagated to all generated npm resource specs.
	RegistryToken string `yaml:",omitempty"`
	// Age defines the minimum or maximum age of a release to be considered valid.
	// It accepts a duration string (e.g., "24h", "7d", "3w", "1y").
	// This will be propagated to all generated npm resource specs.
	//
	// By default, `minimum` is set to `3d` so Updatecli doesn't suggest a package version
	// published less than three days ago. Specifying an empty `age: {}` disables that behavior.
	Age *age.Spec `yaml:",omitempty"`
	// Vulnerability switches the autodiscovery to security updates, based on the OSV database (https://osv.dev).
	// A package is only updated when its current version has known vulnerabilities,
	// to the lowest version without any.
	//
	// example:
	//  ```
	//    vulnerability:
	//      minseverity: high
	//      ignore:
	//        - GHSA-jr5f-v2jv-69x6
	//  ```
	//
	// remark:
	//   * Only security updates are generated, routine updates require a separate manifest.
	//   * It is mutually exclusive with age, versionfilter and ignoreversionconstraints.
	//   * The default minimum release age doesn't apply, a fixed version is suggested as soon as it is known.
	//   * The current version is the exact version from package.json or, for a version constraint,
	//     the version resolved in package-lock.json, pnpm-lock.yaml or yarn.lock next to it.
	//     Packages without an identifiable current version are ignored.
	//   * Labels, such as "security", are set on the action used by the manifest.
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
