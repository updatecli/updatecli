package woodpecker

import (
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec is a struct filled from Updatecli manifest data and shouldn't be modified at runtime unless
// For Fields that requires it, we can use the struct Woodpecker
// Spec defines the parameters which can be provided to the Woodpecker autodiscovery plugin.
type Spec struct {
	// digest provides parameters to specify if the generated manifest should use a digest on top of the tag.
	Digest *bool `yaml:",omitempty"`
	// rootDir defines the root directory used to recursively search for Woodpecker workflow files
	// If rootDir is not provided, the current working directory will be used.
	// If rootDir is provided as an absolute path, scmID will be ignored.
	// If rootDir is not provided but a scmid is, then rootDir will be set to the git repository root directory.
	RootDir string `yaml:",omitempty"`
	// ignore allows to specify rule to ignore autodiscovery a specific Woodpecker workflow based on a rule
	Ignore MatchingRules `yaml:",omitempty"`
	// only allows to specify rule to only autodiscover manifest for a specific Woodpecker workflow based on a rule
	Only MatchingRules `yaml:",omitempty"`
	// auths provides a map of registry credentials where the key is the registry URL without scheme
	// if empty, updatecli relies on OCI credentials such as the one used by Docker.
	//
	// example:
	//
	// ---
	// auths:
	//   "ghcr.io":
	//     token: "xxx"
	//   "index.docker.io":
	//     username: "admin"
	//     password: "password"
	// ---
	//
	Auths map[string]docker.InlineKeyChain `yaml:",omitempty"`
	// FileMatch allows to override default Woodpecker workflow file matching.
	// Default [".woodpecker.yml", ".woodpecker.yaml", ".woodpecker/*.yml", ".woodpecker/*.yaml", ".woodpecker/**/*.yml", ".woodpecker/**/*.yaml"]
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
	//
	// More version filter available at https://www.updatecli.io/docs/core/versionfilter/
	//
	VersionFilter version.Filter `yaml:",omitempty"`
}

// Woodpecker holds all information needed to generate Woodpecker workflow manifest.
type Woodpecker struct {
	// digest holds the value of the digest parameter
	digest bool
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for Woodpecker workflows
	rootDir string
	// filematch defines the filematch rule used to identify Woodpecker workflows that need to be handled
	filematch []string
	// actionID holds the actionID used by the newly generated manifest
	actionID string
	// scmID holds the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid Woodpecker object.
func New(spec interface{}, rootDir, scmID, actionID string) (Woodpecker, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Woodpecker{}, err
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
		return Woodpecker{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, use semantic versioning
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	digest := true
	if s.Digest != nil {
		digest = *s.Digest
	}

	w := Woodpecker{
		actionID:      actionID,
		digest:        digest,
		spec:          s,
		rootDir:       dir,
		filematch:     DefaultFilePatterns,
		scmID:         scmID,
		versionFilter: newFilter,
	}

	if len(s.FileMatch) > 0 {
		w.filematch = s.FileMatch
	}

	return w, nil
}

func (w Woodpecker) DiscoverManifests() ([][]byte, error) {
	logrus.Infof("\n\n%s\n", strings.ToTitle("Woodpecker"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Woodpecker")+1))

	return w.discoverWorkflowImageManifests()
}
