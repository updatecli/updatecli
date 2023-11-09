package updatecli

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines the Updatecli parameters.
type Spec struct {
	// rootdir defines the root directory used to recursively search for Updatecli manifest
	RootDir string `yaml:",omitempty"`
	// Ignore allows to specify rule to ignore "autodiscovery" a specific Updatecli based on a rule
	Ignore MatchingRules `yaml:",omitempty"`
	// Only allows to specify rule to only "autodiscovery" manifest for a specific Updatecli based on a rule
	Only MatchingRules `yaml:",omitempty"`
	// Auths provides a map of registry credentials where the key is the registry URL without scheme
	Auths map[string]docker.InlineKeyChain `yaml:",omitempty"`
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
}

// Updatecli hold all information needed to generate updatecli manifest.
type Updatecli struct {
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootdir defines the root directory from where looking for Updatecli
	rootDir string
	// scmID holds the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid Updatecli object.
func New(spec interface{}, rootDir, scmID string) (Updatecli, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Updatecli{}, err
	}

	dir := rootDir
	if len(s.RootDir) > 0 {
		dir = s.RootDir
	}

	// Fallback to the current process path if no "rootdir" specified.
	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return Updatecli{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, Updatecli policies versioning use semantic versioning
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	return Updatecli{
		spec:          s,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
	}, nil

}

// DiscoverManifests search for Updatecli compose file and generate Updatecli manifests.
func (u Updatecli) DiscoverManifests() ([][]byte, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Updatecli"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Updatecli")+1))

	return u.discoverUpdatecliPolicyManifests()
}
