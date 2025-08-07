package cargo

import (
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/cargo"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines the Cargo parameters.
type Spec struct {
	// RootDir defines the root directory used to recursively search for Cargo.toml
	RootDir string `yaml:",omitempty"`
	// Ignore specifies rule to ignore Cargo.toml update.
	Ignore MatchingRules `yaml:",omitempty"`
	// Only specify required rule to restrict Cargo.toml update.
	Only MatchingRules `yaml:",omitempty"`
	// Auths provides a map of registry credentials where the key is the registry URL without scheme
	Registries map[string]cargo.Registry `yaml:",omitempty"`
	/*
		`versionfilter` provides parameters to specify the version pattern to use when generating manifest.

		kind - semver
			versionfilter of kind `semver` uses semantic versioning as version filtering
			pattern accepts one of:
				`patch` - patch only update patch version
				`minor` - minor only update minor version
				`major` - major only update major versions
				`a version constraint` such as `>= 1.0.0`

		kind - regex
			versionfilter of kind `regex` uses regular expression as version filtering
			pattern accepts a valid regular expression

		example:
		```
			versionfilter:
				kind: semver
				pattern: minor
		```

		and its type like regex, semver, or just latest.
	*/
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
