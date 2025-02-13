package helmfile

import (
	"path"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines the Helmfile parameters.
type Spec struct {
	// rootdir defines the root directory used to recursively search for Helmfile manifest
	RootDir string `yaml:",omitempty"`
	// Ignore allows to specify rule to ignore "autodiscovery" a specific Helmfile based on a rule
	Ignore MatchingRules `yaml:",omitempty"`
	// Only allows to specify rule to only "autodiscovery" manifest for a specific Helmfile based on a rule
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

// Helmfile hold all information needed to generate helmfile manifest.
type Helmfile struct {
	// actionID holds the actionID used by the newly generated manifest
	actionID string
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootdir defines the root directory from where looking for Helmfile
	rootDir string
	// scmID holds the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid Helmfile object.
func New(spec interface{}, rootDir, scmID, actionID string) (Helmfile, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Helmfile{}, err
	}

	dir := rootDir
	if path.IsAbs(s.RootDir) {
		if scmID != "" {
			logrus.Warningf("rootdir %q is an absolute path, scmID %q will be ignored", s.RootDir, scmID)
		}
		dir = s.RootDir
	}

	// Fallback to the current process path if no "rootdir" specified.
	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return Helmfile{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, helmfile release versioning uses semantic versioning
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	return Helmfile{
		actionID:      actionID,
		spec:          s,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
	}, nil

}

func (h Helmfile) DiscoverManifests() ([][]byte, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Helmfile"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Helmfile")+1))

	return h.discoverHelmfileReleaseManifests()
}
