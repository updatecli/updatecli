package helm

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	discoveryConfig "github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery/config"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
)

// Spec defines the Helm parameters.
type Spec struct {
	// rootdir defines the root directory used to recursively search for Helm Chart
	RootDir string `yaml:",omitempty"`
	// Ignore specifies rule to ignore Helm chart update.
	Ignore MatchingRules `yaml:",omitempty"`
	// Only specify required rule to restrict Helm chart update.
	Only MatchingRules `yaml:",omitempty"`
	// Auths provides a map of registry credentials where the key is the registry URL without scheme
	Auths map[string]docker.InlineKeyChain `yaml:",omitempty"`
}

// Helm hold all information needed to generate helm manifest.
type Helm struct {
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootdir defines the root directory from where looking for Helm Chart
	rootDir string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
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

	return Helm{
		spec:    s,
		rootDir: dir,
		scmID:   scmID,
	}, nil

}

func (h Helm) DiscoverManifests(input discoveryConfig.Input) ([]config.Spec, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Helm"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Helm")+1))

	manifests, err := h.discoverHelmDependenciesManifests()

	if err != nil {
		return nil, err
	}

	containerManifest, err := h.discoverHelmContainerManifests()

	if err != nil {
		return nil, err
	}

	manifests = append(manifests, containerManifest...)

	return manifests, nil
}
