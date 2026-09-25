package fleet

import (
	"fmt"
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"rancher/fleet" defines the specification for the Rancher Fleet autodiscovery crawler.
It searches Fleet bundle files, "fleet.yaml" and "fleet.yml", and generates manifests to update the Helm charts they reference.
*/
type Spec struct {
	// "auths" defines the Helm repository credentials, keyed by repository host without scheme.
	//
	// remark:
	//   * only the host part of the repository URL, such as "domain[:port]", is used to look up credentials.
	//
	// example:
	//   ```
	//   auths:
	//     "my-helm-repo.com":
	//       token: "my-secret-token"
	//   ```
	//
	Auths map[string]auth `yaml:",omitempty"`
	// "rootdir" defines the directory where the crawler starts searching for Fleet bundle files.
	//
	// default:
	//   the scm directory when "scmid" is set, otherwise the directory relative paths resolve from, by default the working directory.
	//
	// remark:
	//   * a relative path is resolved from the default directory.
	//   * an absolute path is used as is, instead of the scm directory.
	//
	RootDir string `yaml:",omitempty"`
	// "ignore" defines rules to exclude matching Helm charts from the autodiscovery.
	//
	// remark:
	//   * a Helm chart is ignored when it matches at least one rule.
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// "only" defines rules to restrict the autodiscovery to matching Helm charts.
	//
	// remark:
	//   * a Helm chart is kept only when it matches at least one rule.
	//
	Only MatchingRules `yaml:",omitempty"`
	// "versionfilter" defines the version filter used by the generated manifests.
	//
	// default:
	//   kind "semver" with pattern "*", the latest version.
	//
	// remark:
	//   * with kind "semver", "pattern" accepts:
	//     * "prerelease": the latest prerelease of the current version.
	//     * "patch": patch updates only.
	//     * "minor": patch and minor updates.
	//     * "minoronly": minor updates only.
	//     * "major": patch, minor and major updates.
	//     * "majoronly": major updates only.
	//     * a version constraint, such as ">= 1.0.0".
	//   * with kind "regex", "pattern" accepts a regular expression.
	//   * more examples at https://www.updatecli.io/docs/core/versionfilter/
	//
	// example:
	//   ```
	//   versionfilter:
	//     kind: semver
	//     pattern: minor
	//   ```
	//
	VersionFilter version.Filter `yaml:",omitempty"`
}

// auth defines the credentials used to access a Helm repository.
type auth struct {
	// "token" defines the token used to authenticate with the Helm repository.
	Token string `yaml:",omitempty"`
}

// Fleet holds all information needed to generate fleet bundle.
type Fleet struct {
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for Fleet bundle
	rootDir string
	//actionID hold the actionID used by the newly generated manifest
	actionID string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid Fleet object.
func New(spec interface{}, rootDir, scmID, actionID string) (Fleet, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Fleet{}, err
	}

	// Validate ignore rules
	if err := s.Ignore.Validate(); err != nil {
		return Fleet{}, fmt.Errorf("invalid ignore spec: %w", err)
	}

	// Validate only rules
	if err := s.Only.Validate(); err != nil {
		return Fleet{}, fmt.Errorf("invalid only spec: %w", err)
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
		return Fleet{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, helm versioning uses semantic versioning. Containers is not but...
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	return Fleet{
		spec:          s,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
		actionID:      actionID,
	}, nil

}

func (f Fleet) DiscoverManifests() ([][]byte, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Rancher Fleet"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Rancher Fleet")+1))

	return f.discoverFleetDependenciesManifests()
}
