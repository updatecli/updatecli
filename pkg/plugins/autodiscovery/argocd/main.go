package argocd

import (
	"path"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines the parameters which can be provided to the argocd builder.
type Spec struct {
	// RootDir defines the root directory used to recursively search for ArgoCD manifest
	RootDir string `yaml:",omitempty"`
	// Ignore allows to specify rule to ignore autodiscovery a specific Argocd manifest based on a rule
	Ignore MatchingRules `yaml:",omitempty"`
	// Only allows to specify rule to only autodiscover manifest for a specific ArgoCD manifest based on a rule
	Only MatchingRules `yaml:",omitempty"`
	//  versionfilter provides parameters to specify the version pattern used when generating manifest.
	//
	//  kind - semver
	//    versionfilter of kind `semver` uses semantic versioning as version filtering
	//    pattern accepts one of:
	//      `patch` - patch only update patch version
	//      `minor` - minor only update minor version
	//      `major` - major only update major versions
	//      `a version constraint` such as `>= 1.0.0`
	//
	//  kind - regex
	//    versionfilter of kind `regex` uses regular expression as version filtering
	//    pattern accepts a valid regular expression
	//
	//  example:
	//  ```
	//  	versionfilter:
	//  		kind: semver
	//  		pattern: minor
	//  ```
	//
	//	and its type like regex, semver, or just latest.
	VersionFilter version.Filter `yaml:",omitempty"`
}

// ArgoCD holds all information needed to generate argocd pipelines.
type ArgoCD struct {
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for ArgoCD manifest
	rootDir string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid ArgoCD object.
func New(spec interface{}, rootDir, scmID string) (ArgoCD, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return ArgoCD{}, err
	}

	dir := rootDir

	if path.IsAbs(s.RootDir) {
		if scmID != "" {
			logrus.Warningf("rootdir %q is an absolute path, scmID %q will be ignored", s.RootDir, scmID)
		}
		dir = s.RootDir
	}

	// If no RootDir have been provided via settings,
	// then fallback to the current process path.
	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return ArgoCD{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, helm versioning uses semantic versioning. Containers is not but...
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	return ArgoCD{
		spec:          s,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
	}, nil

}

func (f ArgoCD) DiscoverManifests() ([][]byte, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("ArgoCD"))
	logrus.Infof("%s\n", strings.Repeat("=", len("ArgoCD")+1))

	return f.discoverArgoCDManifests()
}
