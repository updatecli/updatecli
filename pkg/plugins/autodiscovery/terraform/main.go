package terraform

import (
	"path"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Terraform struct holds all information needed to generate terraform manifest.
type Terraform struct {
	// actionID holds the actionID used by the newly generated manifest
	actionID string
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootdir defines the root directory from where looking for .terraform.lock.hcl
	rootDir string
	// scmID holds the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid Terraform object.
func New(spec interface{}, rootDir, scmID, actionID string) (Terraform, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Terraform{}, err
	}

	if len(s.Platforms) == 0 {
		logrus.Warningf("No platforms specified, we fallback to linux_amd64, linux_arm64, darwin_amd64, darwin_arm64")
		s.Platforms = []string{"linux_amd64", "linux_arm64", "darwin_amd64", "darwin_arm64"}
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		logrus.Debugln("no versioning filtering specified, fallback to semantic versioning")
		// By default, golang versioning uses semantic versioning
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	dir := rootDir
	if path.IsAbs(s.RootDir) {
		if scmID != "" {
			logrus.Warningf("rootdir %q is an absolute path, scmID %q will be ignored", s.RootDir, scmID)
		}
		dir = s.RootDir
	}

	// Fallback to the current process path if not rootdir specified.
	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return Terraform{}, err
	}

	return Terraform{
		actionID:      actionID,
		spec:          s,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
	}, nil
}

func (t Terraform) DiscoverManifests() ([][]byte, error) {
	logrus.Infof("\n\n%s\n", strings.ToTitle("Terraform"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Terraform")+1))

	manifests, err := t.discoverTerraformProvidersManifests()
	if err != nil {
		return nil, err
	}

	return manifests, nil
}
