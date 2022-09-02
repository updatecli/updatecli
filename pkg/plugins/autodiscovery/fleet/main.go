package fleet

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	discoveryConfig "github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/pullrequest"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Spec defines the parameters which can be provided to the fleet builder.
type Spec struct {
	// RootDir defines the root directory used to recursively search for Fleet bundle
	RootDir string `yaml:",omitempty"`
	// Disable allows to disable the Fleet crawler
	Disable bool `yaml:",omitempty"`
	// Ignore allows to specify rule to ignore autodiscovery a specific Fleet bundle based on a rule
	Ignore MatchingRules `yaml:",omitempty"`
	// Only allows to specify rule to only autodiscover manifest for a specific Fleet bundle based on a rule
	Only MatchingRules `yaml:",omitempty"`
}

// Fleet hold all information needed to generate fleet bundle.
type Fleet struct {
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for Fleet bundle
	rootDir string
}

// New return a new valid Fleet object.
func New(spec interface{}, rootDir string) (Fleet, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Fleet{}, err
	}

	dir := rootDir
	if len(s.RootDir) > 0 {
		dir = s.RootDir
	}

	// If no RootDir have been provided via settings,
	// then fallback to the current process path.
	if len(dir) == 0 {
		logrus.Errorln("no working directrory defined")
		return Fleet{}, err
	}

	return Fleet{
		spec:    s,
		rootDir: dir,
	}, nil

}

func (f Fleet) DiscoverManifests(input discoveryConfig.Input) ([]config.Spec, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Rancher Fleet"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Rancher Fleet")+1))

	manifests, err := f.discoverFleetDependenciesManifests()

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

// RunDisabled returns a bool saying if a run should be done
func (f Fleet) Enabled() bool {
	return !f.spec.Disable
}
