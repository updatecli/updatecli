package helmfile

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	discoveryConfig "github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery/config"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
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
}

// Helmfile hold all information needed to generate helmfile manifest.
type Helmfile struct {
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootdir defines the root directory from where looking for Helmfile
	rootDir string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
}

// New return a new valid Helmfile object.
func New(spec interface{}, rootDir, scmID string) (Helmfile, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Helmfile{}, err
	}

	dir := rootDir
	if len(s.RootDir) > 0 {
		dir = s.RootDir
	}

	// Fallback to the current process path if no "rootdir" specified.
	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return Helmfile{}, err
	}

	return Helmfile{
		spec:    s,
		rootDir: dir,
		scmID:   scmID,
	}, nil

}

func (h Helmfile) DiscoverManifests(input discoveryConfig.Input) ([]config.Spec, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Helmfile"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Helmfile")+1))

	return h.discoverHelmfileReleaseManifests()
}
