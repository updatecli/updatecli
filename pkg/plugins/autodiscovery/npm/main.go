package npm

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

// Spec defines the parameters which can be provided to the NPM builder.
type Spec struct {
	// RootDir defines the root directory used to recursively search for npm packages.json
	RootDir string `yaml:",omitempty"`
	// Ignore allows to specify rule to ignore autodiscovery a specific NPM based on a rule
	Ignore MatchingRules `yaml:",omitempty"`
	// Only allows to specify rule to only autodiscover manifest for a specific NPM based on a rule
	Only MatchingRules `yaml:",omitempty"`
	// StrictSemver allows to restrict update to identified version matching strict semantic versioning
	StrictSemver bool
}

// Npm holds all information needed to generate npm manifest.
type Npm struct {
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for NPM
	rootDir string
	// scmID holds the scmID used by the newly generated manifest
	scmID string
}

// New return a new valid object.
func New(spec interface{}, rootDir, scmID string) (Npm, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Npm{}, err
	}

	dir := rootDir
	if len(s.RootDir) > 0 {
		dir = s.RootDir
	}

	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return Npm{}, err
	}

	return Npm{
		spec:    s,
		rootDir: dir,
		scmID:   scmID,
	}, nil

}

func (n Npm) DiscoverManifests() ([][]byte, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("NPM"))
	logrus.Infof("%s\n", strings.Repeat("=", len("NPM")+1))

	manifests, err := n.discoverDependencyManifests()

	if err != nil {
		return nil, err
	}

	return manifests, nil
}
