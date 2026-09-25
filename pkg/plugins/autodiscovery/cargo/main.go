package cargo

import (
	"fmt"
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/cargo"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"cargo" defines the specification for the Cargo autodiscovery crawler.
It searches "Cargo.toml" files and generates manifests to update the crate dependencies they declare.
*/
type Spec struct {
	// "rootdir" defines the directory where the crawler starts searching for "Cargo.toml" files.
	//
	// default:
	//   the scm directory when "scmid" is set, otherwise the directory relative paths resolve from, by default the working directory.
	//
	// remark:
	//   * a relative path is resolved from the default directory.
	//   * an absolute path is used as is, instead of the scm directory.
	//
	RootDir string `yaml:",omitempty"`
	// "ignore" defines rules to exclude matching crates from the autodiscovery.
	//
	// remark:
	//   * a crate is ignored when it matches at least one rule.
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// "only" defines rules to restrict the autodiscovery to matching crates.
	//
	// remark:
	//   * a crate is kept only when it matches at least one rule.
	//
	Only MatchingRules `yaml:",omitempty"`
	// "registries" defines the Cargo registries used by the generated manifests, keyed by registry name.
	//
	// remark:
	//   * the key is the name set in the "registry" key of a dependency in "Cargo.toml".
	//
	Registries map[string]cargo.Registry `yaml:",omitempty"`
	// "versionfilter" defines the version filter used by the generated manifests.
	//
	// default:
	//   kind "semver" with pattern "*", the latest version.
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
	//   * with kind "semver" and a pattern other than "*", a dependency whose version is not a strict semantic version uses its declared version as the pattern.
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

// Cargo struct holds all information needed to generate cargo manifest.
type Cargo struct {
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootdir defines the root directory from where looking for Helm Chart
	rootDir string
	//actionID hold the actionID used by the newly generated manifest
	actionID string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	// cargoAvailable tells if `cargo` is available
	cargoAvailable bool
	// cargoUpgradeAvailable tells if `cargo upgrade` (from cargo-edits) is available
	cargoUpgradeAvailable bool
}

// New return a new valid Cargo object.
func New(spec interface{}, rootDir, scmID, actionID string) (Cargo, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Cargo{}, err
	}

	// Validate ignore rules
	if err := s.Ignore.Validate(); err != nil {
		return Cargo{}, fmt.Errorf("invalid ignore spec: %w", err)
	}

	// Validate only rules
	if err := s.Only.Validate(); err != nil {
		return Cargo{}, fmt.Errorf("invalid only spec: %w", err)
	}

	dir := rootDir
	if path.IsAbs(s.RootDir) {
		if scmID != "" {
			logrus.Warningf("rootdir %q is an absolute path, scmID %q will be ignored", s.RootDir, scmID)
		}
		dir = s.RootDir
	}

	// Fallback to the current process path if not rootdir specified.
	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return Cargo{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		logrus.Debugln("no versioning filtering specified, fallback to semantic versioning")
		// By default, golang versioning uses semantic versioning
		newFilter.Kind = version.SEMVERVERSIONKIND
		newFilter.Pattern = "*"
	}

	return Cargo{
		actionID:              actionID,
		spec:                  s,
		rootDir:               dir,
		scmID:                 scmID,
		versionFilter:         newFilter,
		cargoAvailable:        isCargoAvailable(),
		cargoUpgradeAvailable: isCargoUpgradeAvailable(),
	}, nil

}

func (c Cargo) DiscoverManifests() ([][]byte, error) {
	logrus.Infof("\n\n%s\n", strings.ToTitle("Cargo"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Cargo")+1))

	manifests, err := c.discoverCargoDependenciesManifests()
	if err != nil {
		return nil, err
	}

	return manifests, nil
}
