package cargo

import (
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/cargo"
	"strings"
)

// Spec defines the Cargo parameters.
type Spec struct {
	// rootdir defines the root directory used to recursively search for Cargo.toml
	RootDir string `yaml:",omitempty"`
	// Ignore specifies rule to ignore Cargo.toml update.
	Ignore MatchingRules `yaml:",omitempty"`
	// Only specify required rule to restrict Cargo.toml update.
	Only MatchingRules `yaml:",omitempty"`
	// Auths provides a map of registry credentials where the key is the registry URL without scheme
	Auths map[string]cargo.InlineKeyChain `yaml:",omitempty"`
}

// Cargo struct holds all information needed to generate cargo manifest.
type Cargo struct {
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootdir defines the root directory from where looking for Helm Chart
	rootDir string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
}

// New return a new valid Cargo object.
func New(spec interface{}, rootDir, scmID string) (Cargo, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Cargo{}, err
	}

	dir := rootDir
	if len(s.RootDir) > 0 {
		dir = s.RootDir
	}

	// Fallback to the current process path if not rootdir specified.
	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return Cargo{}, err
	}

	return Cargo{
		spec:    s,
		rootDir: dir,
		scmID:   scmID,
	}, nil

}

func (c Cargo) DiscoverManifests() ([][]byte, error) {
	logrus.Infof("\n\n%s\n", strings.ToTitle("Cargo"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Cargo")+1))

	manifests, err := c.discoverCargoDependenciesManifests()
	if err != nil {
		return nil, err
	}

	return manifests, nil
}
