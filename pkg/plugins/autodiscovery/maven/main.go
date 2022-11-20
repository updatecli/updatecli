package maven

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"

	discoveryConfig "github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery/config"
)

// Spec defines the parameters which can be provided to the Helm builder.
type Spec struct {
	// RootDir defines the root directory used to recursively search for Helm Chart
	RootDir string `yaml:",omitempty"`
	// Ignore allows to specify rule to ignore autodiscovery a specific Helm based on a rule
	Ignore MatchingRules `yaml:",omitempty"`
	// Only allows to specify rule to only autodiscover manifest for a specific Helm based on a rule
	Only MatchingRules `yaml:",omitempty"`
}

// Maven hold all information needed to generate helm manifest.
type Maven struct {
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for Helm Chart
	rootDir string
}

// New return a new valid Helm object.
func New(spec interface{}, rootDir string) (Maven, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Maven{}, err
	}

	dir := rootDir
	if len(s.RootDir) > 0 {
		dir = s.RootDir
	}

	// If no RootDir have been provided via settings,
	// then fallback to the current process path.
	if len(dir) == 0 {
		logrus.Errorln("no working directrory defined")
		return Maven{}, err
	}

	return Maven{
		spec:    s,
		rootDir: dir,
	}, nil

}

func (m Maven) DiscoverManifests(input discoveryConfig.Input) ([]config.Spec, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Maven"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Maven")+1))

	manifests, err := m.discoverDependenciesManifests()

	if err != nil {
		return nil, err
	}

	dependencyManagementManifests, err := m.discoverDependencyManagementsManifests()

	if err != nil {
		return nil, err
	}

	manifests = append(manifests, dependencyManagementManifests...)

	parentPomdependencyManifests, err := m.discoverParentPomDependencyManifests()

	if err != nil {
		return nil, err
	}

	manifests = append(manifests, parentPomdependencyManifests...)

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
