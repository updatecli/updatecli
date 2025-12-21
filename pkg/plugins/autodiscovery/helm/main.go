package helm

import (
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines the Helm parameters.
type Spec struct {
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
	// digest provides a parameter to specify if the generated manifest should use a digest on top of the tag when updating container.
	Digest *bool `yaml:",omitempty"`
	// ignorecontainer disables OCI container tag update when set to true
	IgnoreContainer bool `yaml:",omitempty"`
	// ignorechartdependency disables Helm chart dependencies update when set to true
	IgnoreChartDependency bool `yaml:",omitempty"`
	// Ignore specifies rule to ignore Helm chart update.
	Ignore MatchingRules `yaml:",omitempty"`
	// rootdir defines the root directory used to recursively search for Helm Chart
	RootDir string `yaml:",omitempty"`
	// only specify required rule(s) to restrict Helm chart update.
	Only MatchingRules `yaml:",omitempty"`
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
	// [target] Defines if a Chart should be packaged or not.
	SkipPackaging bool `yaml:",omitempty"`
	// [target] Defines if a Chart changes, triggers, or not, a Chart version update, accepted values is a comma separated list of "none,major,minor,patch"
	VersionIncrement string `yaml:",omitempty"`
}

// Helm hold all information needed to generate helm manifest.
type Helm struct {
	// digest holds the value of the digest parameter
	digest bool
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootdir defines the root directory from where looking for Helm Chart
	rootDir string
	// actionID hold the actionID used by the newly generated manifest
	actionID string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid Helm object.
func New(spec interface{}, rootDir, scmID, actionID string) (Helm, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Helm{}, err
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
		return Helm{}, err
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

	return Helm{
		actionID:      actionID,
		digest:        digest,
		spec:          s,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
	}, nil

}

func (h Helm) DiscoverManifests() ([][]byte, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Helm"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Helm")+1))

	var manifests [][]byte
	var err error

	if !h.spec.IgnoreChartDependency {
		manifests, err = h.discoverHelmDependenciesManifests()
		if err != nil {
			logrus.Errorf("generating Helm chart dependencies manifest(s): %s", err)
		}
	}

	if !h.spec.IgnoreContainer {
		containerManifest, err := h.discoverHelmContainerManifests()
		if err != nil {
			logrus.Errorf("generating container update manifest(s): %s", err)
		}

		manifests = append(manifests, containerManifest...)
	}

	return manifests, nil
}
