package golang

import (
	"fmt"
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines the parameters which can be provided to the Golang autodiscovery builder.
type Spec struct {
	// rootDir defines the root directory used to recursively search for golang go.mod
	RootDir string `yaml:",omitempty"`
	// OnlyGoVersion allows to specify if the autodiscovery should only handle Go version specified in go.mod
	OnlyGoVersion *bool `yaml:",omitempty"`
	// OnlyGoModule allows to specify if the autodiscovery should only handle Go module specified in go.mod
	OnlyGoModule *bool `yaml:",omitempty"`
	// ignore allows to specify `rule` to ignore autodiscovery a specific go.mod rule
	Ignore MatchingRules `yaml:",omitempty"`
	// `only` allows to specify rule to `only` autodiscover manifest for a specific golang rule
	Only MatchingRules `yaml:",omitempty"`
	//  `versionfilter` provides parameters to specify the version pattern used when generating manifest.
	//
	//  kind - semver
	//    versionfilter of kind `semver` uses semantic versioning as version filtering
	//    pattern accepts one of:
	//      `prerelease` - Updatecli tries to identify the latest prerelease whatever it means
	//      `patch` - Updatecli only handles patch version update
	//      `minor` - Updatecli handles patch AND minor version update
	//      `minoronly` - Updatecli handles minor version only
	//      `major` - Updatecli handles patch, minor, AND major version update
	//      `majoronly` - Updatecli only handles major version update
	//      `a version constraint` such as `>= 1.0.0`
	//
	//  kind - regex
	//    versionfilter of kind `regex` uses regular expression as version filtering
	//    pattern accepts a valid regular expression
	//
	//  example:
	//  ```
	//    versionfilter:
	//      kind: semver
	//      pattern: minor
	//  ```
	//
	//  and its type like regex, semver, or just latest.
	//
	//  More examples can be found at https://www.updatecli.io/docs/core/versionfilter/
	VersionFilter version.Filter `yaml:",omitempty"`
	// Age defines the minimum or maximum age of a release to be considered valid.
	// It accepts a duration string (e.g., `24h`, `7d`, `1w`).
	Age age.Spec `yaml:",omitempty"`
}

// Golang holds all information needed to generate golang manifest.
type Golang struct {
	onlyGoModule  bool
	onlygoVersion bool
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for go.mod file
	rootDir string
	// actionID holds the actionID used by the newly generated manifest
	actionID string
	// scmID holds the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid object.
func New(spec interface{}, rootDir, scmID, actionID string) (Golang, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Golang{}, err
	}

	// Validate ignore rules
	if err := s.Ignore.Validate(); err != nil {
		return Golang{}, fmt.Errorf("invalid ignore spec: %w", err)
	}

	// Validate only rules
	if err := s.Only.Validate(); err != nil {
		return Golang{}, fmt.Errorf("invalid only spec: %w", err)
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		logrus.Debugln("no versioning filtering specified, fallback to semantic versioning")
		// By default, golang versioning uses semantic versioning
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	goVersionOnly := false
	if s.OnlyGoVersion != nil {
		goVersionOnly = *s.OnlyGoVersion
	}

	goModuleOnly := false
	if s.OnlyGoModule != nil {
		goModuleOnly = *s.OnlyGoModule
	}

	dir := rootDir
	if path.IsAbs(s.RootDir) {
		if scmID != "" {
			logrus.Warningf("rootdir %q is an absolute path, scmID %q will be ignored", s.RootDir, scmID)
		}
		dir = s.RootDir
	}

	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return Golang{}, err
	}

	g := Golang{
		spec:          s,
		rootDir:       dir,
		scmID:         scmID,
		onlyGoModule:  goModuleOnly,
		onlygoVersion: goVersionOnly,
		versionFilter: newFilter,
		actionID:      actionID,
	}

	return g, nil

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
