package dockerfile

import (
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec is a struct fill from Updatecli manifest data and shouldn't be modified at runtime unless
// For Fields that requires it, we can use the struct Dockerfile
// Spec defines the parameters which can be provided to the Dockerfile crawler.
type Spec struct {
	/*
		digest provides parameters to specify if the generated manifest should use a digest on top of the tag.
	*/
	Digest *bool `yaml:",omitempty"`
	// RootDir defines the root directory used to recursively search for Helm Chart
	RootDir string `yaml:",omitempty"`
	// Ignore allows to specify rule to ignore autodiscovery a specific Helm based on a rule
	Ignore MatchingRules `yaml:",omitempty"`
	// Only allows to specify rule to only autodiscover manifest for a specific Helm based on a rule
	Only MatchingRules `yaml:",omitempty"`
	// Auths provides a map of registry credentials where the key is the registry URL without scheme
	Auths map[string]docker.InlineKeyChain `yaml:",omitempty"`
	// FileMatch allows to override default Dockerfile file matching. Default ["Dockerfile"]
	FileMatch []string `yaml:",omitempty"`
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

// Dockerfile hold all information needed to generate Dockerfile manifest.
type Dockerfile struct {
	// digest holds the value of the digest parameter
	digest bool
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for Helm Chart
	rootDir string
	// filematch defines the filematch rule used to identify the Dockerfile that need to be handled
	filematch []string
	// actionID hold the actionID used by the newly generated manifest
	actionID string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid Helm object.
func New(spec interface{}, rootDir, scmID, actionID string) (Dockerfile, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Dockerfile{}, err
	}

	dir := rootDir
	if path.IsAbs(s.RootDir) {
		if scmID != "" {
			logrus.Warningf("rootdir %q is an absolute path, scmID %q will be ignored", s.RootDir, scmID)
		}
		dir = s.RootDir
	}

	// If no RootDir have been provided via settings,
	// then fallback to the current process path.
	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return Dockerfile{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, helm versioning uses semantic versioning. Containers is not but...
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	digest := true
	if s.Digest != nil {
		digest = *s.Digest
	}

	d := Dockerfile{
		actionID:      actionID,
		digest:        digest,
		spec:          s,
		rootDir:       dir,
		filematch:     DefaultFileMatch,
		scmID:         scmID,
		versionFilter: newFilter,
	}

	if len(s.FileMatch) > 0 {
		d.filematch = s.FileMatch
	}

	return d, nil

}

func (d Dockerfile) DiscoverManifests() ([][]byte, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Dockerfile"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Dockerfile")+1))

	return d.discoverDockerfileManifests()
}
