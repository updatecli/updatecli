package nomad

import (
	"path"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec is a struct fill from Updatecli manifest data and shouldn't be modified at runtime unless
// For Fields that requires it, we can use the struct DockerCompose
// Spec defines the parameters which can be provided to the Helm builder.
type Spec struct {
	// digest provides parameters to specify if the generated manifest should use a digest on top of the tag.
	Digest *bool `yaml:",omitempty"`
	// rootDir defines the root directory used to recursively search for Helm Chart
	// If rootDir is not provided, the current working directory will be used.
	// If rootDir is provided as an absolute path, scmID will be ignored.
	// If rootDir is not provided but a scmid is, then rootDir will be set to the git repository root directory.
	RootDir string `yaml:",omitempty"`
	// ignore allows to specify rule to ignore autodiscovery a specific Helm based on a rule
	Ignore MatchingRules `yaml:",omitempty"`
	// only allows to specify rule to only autodiscover manifest for a specific Helm based on a rule
	Only MatchingRules `yaml:",omitempty"`
	// auths provides a map of registry credentials where the key is the registry URL without scheme
	Auths map[string]docker.InlineKeyChain `yaml:",omitempty"`
	// FileMatch allows to override default docker-compose.yaml file matching. Default ["docker-compose.yaml","docker-compose.yml","docker-compose.*.yaml","docker-compose.*.yml"]
	FileMatch []string `yaml:",omitempty"`
	// versionfilter provides parameters to specify the version pattern used when generating manifest.
	//
	// More information available at
	// https://www.updatecli.io/docs/core/versionfilter/
	//
	// kind - semver
	//   versionfilter of kind `semver` uses semantic versioning as version filtering
	//   pattern accepts one of:
	//     `patch` - patch only update patch version
	//     `minor` - minor only update minor version
	//     `major` - major only update major versions
	//     `a version constraint` such as `>= 1.0.0`
	//
	// kind - regex
	// versionfilter of kind `regex` uses regular expression as version filtering
	// pattern accepts a valid regular expression
	//
	// example:
	// ```
	//   versionfilter:
	//   kind: semver
	//   pattern: minor
	//```
	//and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
}

// Nomad hold all information needed to generate Updatecli manifest to update Nomad configuration.
// Today we only support autodiscovering docker image tag and digest
type Nomad struct {
	// digest holds the value of the digest parameter
	digest bool
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for Nomad files
	// If rootDir is not provided, the current working directory will be used.
	// If a scmid is provided and rootDir is not provided, then rootDir will be set to the git repository root directory.
	rootDir string
	// filematch defines the filematch rule used to identify Nomad files that need to be handled
	filematch []string
	// actionID holds the actionID used by the newly generated manifest
	actionID string
	// scmID holds the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid Nomad object.
func New(spec interface{}, rootDir, scmID, actionID string) (Nomad, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Nomad{}, err
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
		return Nomad{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, NPM versioning uses semantic versioning
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	digest := true
	if s.Digest != nil {
		digest = *s.Digest
	}

	n := Nomad{
		actionID:      actionID,
		digest:        digest,
		spec:          s,
		rootDir:       dir,
		filematch:     DefaultFilePattern,
		scmID:         scmID,
		versionFilter: newFilter,
	}

	if len(s.FileMatch) > 0 {
		n.filematch = s.FileMatch
	}

	return n, nil

}

func (n Nomad) DiscoverManifests() ([][]byte, error) {
	logrus.Infof("\n\n%s\n", strings.ToTitle("Nomad"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Nomad")+1))

	return n.discoverDockerDriverManifests()
}
