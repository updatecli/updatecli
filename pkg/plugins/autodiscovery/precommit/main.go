package precommit

import (
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines the parameters uses to generate the precomit manifests
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
	// Digest provides parameters to specify if the generated manifest should use a digest instead of the branch or tag.
	// This is equivalent to using the [`--freeze`](https://pre-commit.com/#pre-commit-autoupdate) option of precommit
	Digest *bool `yaml:",omitempty"`
}

// Precommit holds all information needed to generate precommit manifest.
type Precommit struct {
	// actionID holds the actionID used by the newly generated manifest
	actionID string
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for precommit
	rootDir string
	// scmID holds the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	// digest holds the value of the digest parameter
	digest bool
}

// New return a new valid object.
func New(spec interface{}, rootDir, scmID, actionID string) (Precommit, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Precommit{}, err
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
		return Precommit{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, Updatecli policies versioning use semantic versioning
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	digest := false
	if s.Digest != nil {
		digest = *s.Digest
	}

	return Precommit{
		actionID:      actionID,
		spec:          s,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
		digest:        digest,
	}, nil

}

func (p Precommit) DiscoverManifests() ([][]byte, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("PRECOMMIT"))
	logrus.Infof("%s\n", strings.Repeat("=", len("PRECOMMIT")+1))

	manifests, err := p.discoverDependencyManifests()

	if err != nil {
		return nil, err
	}

	return manifests, nil
}
