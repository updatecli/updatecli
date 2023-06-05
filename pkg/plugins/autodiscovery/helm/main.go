package helm

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines the Helm parameters.
type Spec struct {
	/*
		auths provides a map of registry credentials where the key is the registry URL without scheme
		if empty, updatecli relies on OCI credentials such as the one used by Docker.
	*/
	Auths map[string]docker.InlineKeyChain `yaml:",omitempty"`
	// ignorecontainer disables OCI container tag update when set to true
	IgnoreContainer bool `yaml:",omitempty"`
	// ignorechartdepency disables Helm chart dependencies update when set to true
	IgnoreChartDependency bool `yaml:",omitempty"`
	// Ignore specifies rule to ignore Helm chart update.
	Ignore MatchingRules `yaml:",omitempty"`
	// rootdir defines the root directory used to recursively search for Helm Chart
	RootDir string `yaml:",omitempty"`
	// only specify required rule(s) to restrict Helm chart update.
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
}

// Helm hold all information needed to generate helm manifest.
type Helm struct {
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootdir defines the root directory from where looking for Helm Chart
	rootDir string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid Helm object.
func New(spec interface{}, rootDir, scmID string) (Helm, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Helm{}, err
	}

	dir := rootDir
	if len(s.RootDir) > 0 {
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

	return Helm{
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
