package dockerfile

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

// Spec is a struct fill from Updatecli manifest data and shouldn't be modified at runtime unless
// For Fields that requires it, we can use the struct Dockerfile
// Spec defines the parameters which can be provided to the Dockerfile crawler.
type Spec struct {
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
}

// Dockerfile hold all information needed to generate Dockerfile manifest.
type Dockerfile struct {
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for Helm Chart
	rootDir string
	// filematch defines the filematch rule used to identify the Dockerfile that need to be handled
	filematch []string
}

// New return a new valid Helm object.
func New(spec interface{}, rootDir string) (Dockerfile, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Dockerfile{}, err
	}

	dir := rootDir
	if len(s.RootDir) > 0 {
		dir = s.RootDir
	}

	// If no RootDir have been provided via settings,
	// then fallback to the current process path.
	if len(dir) == 0 {
		logrus.Errorln("no working directrory defined")
		return Dockerfile{}, err
	}

	d := Dockerfile{
		spec:      s,
		rootDir:   dir,
		filematch: DefaultFileMatch,
	}

	if len(s.FileMatch) > 0 {
		d.filematch = s.FileMatch
	}

	return d, nil

}

func (h Dockerfile) DiscoverManifests(input discoveryConfig.Input) ([]config.Spec, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Dockerfile"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Dockerfile")+1))

	manifests, err := h.discoverDockerfileManifests()

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
