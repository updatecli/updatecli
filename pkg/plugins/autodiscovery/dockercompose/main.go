package dockercompose

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
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

// DockerCompose hold all information needed to generate compose file manifest.
type DockerCompose struct {
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for Helm Chart
	rootDir string
	// filematch defines the filematch rule used to identify docker-compose that need to be handled
	filematch []string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
}

// New return a new valid Helm object.
func New(spec interface{}, rootDir, scmID string) (DockerCompose, error) {
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
		logrus.Errorln("no working directory defined")
		return DockerCompose{}, err
	}

	d := DockerCompose{
		spec:      s,
		rootDir:   dir,
		filematch: []string{DefaultFilePattern},
		scmID:     scmID,
	}

	if len(s.FileMatch) > 0 {
		d.filematch = s.FileMatch
	}

	return d, nil

}

func (h DockerCompose) DiscoverManifests() ([][]byte, error) {
	// Print the header to get started
	logrus.Infof("\n\n%s\n", strings.ToTitle("Docker Compose"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Docker Compose")+1))

	return h.discoverDockerComposeImageManifests()
}
