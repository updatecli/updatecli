package dockercompose

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	discoveryConfig "github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/pullrequest"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
)

// Spec is a struct fill from Updatecli manifest data and shouldn't be modified at runtime unless
// For Fields that requires it, we can use the struct DockerCompose
// Spec defines the parameters which can be provided to the Helm builder.
type Spec struct {
	// RootDir defines the root directory used to recursively search for Helm Chart
	RootDir string `yaml:",omitempty"`
	// Ignore allows to specify rule to ignore autodiscovery a specific Helm based on a rule
	Ignore MatchingRules `yaml:",omitempty"`
	// Only allows to specify rule to only autodiscover manifest for a specific Helm based on a rule
	Only MatchingRules `yaml:",omitempty"`
	// Auths provides a map of registry credentials where the key is the registry URL without scheme
	Auths map[string]docker.InlineKeyChain `yaml:",omitempty"`
	// FileMatch allows to override default docker-compose.yaml file matching. Default ["docker-compose.yaml","docker-compose.yml","docker-compose.*.yaml","docker-compose.*.yml"]
	FileMatch []string `yaml:",omitempty"`
}

// DockerCompose hold all information needed to generate helmfile manifest.
type DockerCompose struct {
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for Helm Chart
	rootDir string
	// filematch defines the filematch rule used to identify docker-compose that need to be handled
	filematch []string
}

// New return a new valid Helm object.
func New(spec interface{}, rootDir string) (DockerCompose, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return DockerCompose{}, err
	}

	dir := rootDir
	if len(s.RootDir) > 0 {
		dir = s.RootDir
	}

	// If no RootDir have been provided via settings,
	// then fallback to the current process path.
	if len(dir) == 0 {
		logrus.Errorln("no working directrory defined")
		return DockerCompose{}, err
	}

	d := DockerCompose{
		spec:      s,
		rootDir:   dir,
		filematch: DefaultFileMatch,
	}

	if len(s.FileMatch) > 0 {
		d.filematch = s.FileMatch
	}

	return d, nil

}

func (h DockerCompose) DiscoverManifests(input discoveryConfig.Input) ([]config.Spec, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Docker Compose"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Docker Compose")+1))

	manifests, err := h.discoverDockerComposeImageManifests()

	if err != nil {
		return nil, err
	}

	// Set scm configuration if specified
	for i := range manifests {
		// Set scm configuration if specified
		if len(input.ScmID) > 0 {
			SetScm(&manifests[i], *input.ScmSpec, input.ScmID)
		}

		// Set pullrequest configuration if specified
		if len(input.PullrequestID) > 0 {
			SetPullrequest(&manifests[i], *input.PullRequestSpec, input.PullrequestID)
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

func SetPullrequest(configSpec *config.Spec, pullrequestSpec pullrequest.Config, pullrequestID string) {
	configSpec.PullRequests = make(map[string]pullrequest.Config)
	configSpec.PullRequests[pullrequestID] = pullrequestSpec
}
