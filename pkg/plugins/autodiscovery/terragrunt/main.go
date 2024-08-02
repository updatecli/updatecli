package terragrunt

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Terragrunt struct holds all information needed to generate terragrunt manifest.
type Terragrunt struct {
	// `spec` defines the settings provided via an updatecli manifest
	spec Spec
	// `rootdir` defines the root directory from where looking for terragrunt configuration
	rootDir string
	// `scmID` holds the scmID used by the newly generated manifest
	scmID string
	// `versionFilter` holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid Terragrunt object.
func New(spec interface{}, rootDir, scmID string) (Terragrunt, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Terragrunt{}, err
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

	// Fallback to the current process path if not rootdir specified.
	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return Terragrunt{}, err
	}

	return Terragrunt{
		spec:          s,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
	}, nil
}

func (t Terragrunt) DiscoverManifests() ([][]byte, error) {
	logrus.Infof("\n\n%s\n", strings.ToTitle("Terragrunt"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Terragrunt")+1))

	manifests, err := t.discoverTerragruntManifests()
	if err != nil {
		return nil, err
	}

	return manifests, nil
}
