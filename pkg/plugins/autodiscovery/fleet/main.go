package fleet

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	discoveryConfig "github.com/updatecli/updatecli/pkg/core/pipeline/autodiscovery/config"
)

// Spec defines the parameters which can be provided to the fleet builder.
type Spec struct {
	// RootDir defines the root directory used to recursively search for Fleet bundle
	RootDir string `yaml:",omitempty"`
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
	// scmID hold the scmID used by the newly generated manifest
	scmID string
}

// New return a new valid Fleet object.
func New(spec interface{}, rootDir, scmID string) (Fleet, error) {
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
		scmID:   scmID,
	}, nil

}

func (f Fleet) DiscoverManifests(input discoveryConfig.Input) ([]config.Spec, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Rancher Fleet"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Rancher Fleet")+1))

	return f.discoverFleetDependenciesManifests()
}
