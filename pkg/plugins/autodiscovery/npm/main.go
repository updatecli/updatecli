package npm

import (
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines the parameters which can be provided to the NPM builder.
type Spec struct {
	// RootDir defines the root directory used to recursively search for npm packages.json
	RootDir string `yaml:",omitempty"`
	// Ignore allows to specify rule to ignore autodiscovery a specific NPM based on a rule
	Ignore MatchingRules `yaml:",omitempty"`
	// Only allows to specify rule to only autodiscover manifest for a specific NPM based on a rule
	Only MatchingRules `yaml:",omitempty"`
	/*
		versionfilter provides parameters to specify the version pattern used when generating manifest.

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
	// IgnoreVersionConstraints indicates whether to respect version constraints defined in package.json or not.
	// When set to true, Updatecli will ignore version constraints and update to the latest version available
	// in the registry according to the specified version filter.
	// Default is false.
	//
	// Remark:
	//  * If set to false, Updatecli will try to convert version constrains to valid semantic version
	//    so we can use versionFilter to retrieve the last Major/Minor/Patch version but in case of complex version constraints, such as ">=1.0.0 <2.0.0",
	//    Updatecli will convert it to the first version it detects such as 1.0.0 in our example
	IgnoreVersionConstraints *bool `yaml:",omitempty"`
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
}

// New return a new valid object.
func New(spec interface{}, rootDir, scmID, actionID string) (Npm, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Npm{}, err
	}

	// By default we want to suggest the latest version available in the registry
	ignoreVersionConstraint := true
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
	}, nil

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
