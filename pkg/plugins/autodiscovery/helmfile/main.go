package helmfile

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	discoveryConfig "github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
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
}

// New return a new valid Helmfile object.
func New(spec interface{}, rootDir string) (Helmfile, error) {
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
	}, nil

}

func (h Helmfile) DiscoverManifests(input discoveryConfig.Input) ([]config.Spec, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Helmfile"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Helmfile")+1))

	manifests, err := h.discoverHelmfileReleaseManifests()

	if err != nil {
		return nil, err
	}

	// Set scm configuration if specified
	for i := range manifests {
		// Set scm configuration if specified
		if len(input.ScmID) > 0 {
			SetScm(&manifests[i], *input.ScmSpec, input.ScmID)
		}

		// Set action configuration if specified
		if len(input.ActionID) > 0 {
			SetAction(&manifests[i], *input.ActionConfig, input.ActionID)
		}
	}

	return manifests, nil
}

func SetScm(configSpec *config.Spec, scmSpec scm.Config, scmID string) {
	configSpec.SCMs = make(map[string]scm.Config)
	configSpec.SCMs[scmID] = scmSpec

	for id, condition := range configSpec.Conditions {
		condition.SCMID = scmID
		configSpec.Conditions[id] = condition
	}

	for id, target := range configSpec.Targets {
		target.SCMID = scmID
		configSpec.Targets[id] = target
	}
}

func SetAction(configSpec *config.Spec, actionSpec action.Config, actionID string) {
	configSpec.Actions = make(map[string]action.Config)
	configSpec.Actions[actionID] = actionSpec
}
