package golang

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines the parameters which can be provided to the Golang autodiscovery builder.
type Spec struct {
	// rootDir defines the root directory used to recursively search for golang go.mod
	RootDir string `yaml:",omitempty"`
	// ignore allows to specify rule to ignore autodiscovery a specific go.mod rule
	Ignore MatchingRules `yaml:",omitempty"`
	// only allows to specify rule to "only" autodiscover manifest for a specific golang rule
	Only MatchingRules `yaml:",omitempty"`
	// [S] VersionFilter provides parameters to specify version pattern and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
}

// Golang holds all information needed to generate golang manifest.
type Golang struct {
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for go.mod file
	rootDir string
	// scmID holds the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid object.
func New(spec interface{}, rootDir, scmID string) (Golang, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Golang{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		logrus.Debugln("no versioning filtering specified, fallback to semantic versioning")
		// By default, golang versioning uses semantic versioning
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	dir := rootDir
	if len(s.RootDir) > 0 {
		dir = s.RootDir
	}

	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return Golang{}, err
	}

	return Golang{
		spec:          s,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
	}, nil

}

func (n Golang) DiscoverManifests() ([][]byte, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Golang"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Golang")+1))

	manifests, err := n.discoverDependencyManifests()

	if err != nil {
		return nil, err
	}

	return manifests, nil
}
