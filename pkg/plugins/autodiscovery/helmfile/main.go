package helmfile

import (
	"fmt"
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"helmfile" defines the specification for the Helmfile autodiscovery crawler.
It searches Helmfile files and generates manifests to update their Helm chart releases.
*/
type Spec struct {
	// "rootdir" defines the directory where the crawler starts searching for Helmfile files.
	//
	// default:
	//   the scm directory when "scmid" is set, otherwise the directory relative paths resolve from, by default the working directory.
	//
	// remark:
	//   * a relative path is resolved from the default directory.
	//   * an absolute path is used as is, instead of the scm directory.
	//
	RootDir string `yaml:",omitempty"`
	// "ignore" defines rules to exclude matching releases from the autodiscovery.
	//
	// remark:
	//   * a release is ignored when it matches at least one rule.
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// "only" defines rules to restrict the autodiscovery to matching releases.
	//
	// remark:
	//   * a release is kept only when it matches at least one rule.
	//
	Only MatchingRules `yaml:",omitempty"`
	// "auths" defines the Helm repository credentials, keyed by repository host without scheme.
	//
	// remark:
	//   * only "token" is used.
	//
	// example:
	//   ```
	//   auths:
	//     "my-helm-repo.com":
	//       token: "xxx"
	//   ```
	//
	Auths map[string]docker.InlineKeyChain `yaml:",omitempty"`
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

// Helmfile hold all information needed to generate helmfile manifest.
type Helmfile struct {
	// actionID holds the actionID used by the newly generated manifest
	actionID string
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootdir defines the root directory from where looking for Helmfile
	rootDir string
	// scmID holds the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid Helmfile object.
func New(spec interface{}, rootDir, scmID, actionID string) (Helmfile, error) {
	var s Spec

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		ErrorUnused: true,
		Result:      &s,
	})
	if err != nil {
		return Helmfile{}, err
	}

	if err := decoder.Decode(spec); err != nil {
		return Helmfile{}, fmt.Errorf("invalid spec: %w", err)
	}

	// Validate ignore rules
	if err := s.Ignore.Validate(); err != nil {
		return Helmfile{}, fmt.Errorf("invalid ignore spec: %w", err)
	}

	// Validate only rules
	if err := s.Only.Validate(); err != nil {
		return Helmfile{}, fmt.Errorf("invalid only spec: %w", err)
	}

	dir := rootDir
	if path.IsAbs(s.RootDir) {
		if scmID != "" {
			logrus.Warningf("rootdir %q is an absolute path, scmID %q will be ignored", s.RootDir, scmID)
		}
		dir = s.RootDir
	}

	// Fallback to the current process path if no "rootdir" specified.
	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return Helmfile{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, helmfile release versioning uses semantic versioning
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	return Helmfile{
		actionID:      actionID,
		spec:          s,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
	}, nil

}

func (h Helmfile) DiscoverManifests() ([][]byte, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Helmfile"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Helmfile")+1))

	return h.discoverHelmfileReleaseManifests()
}
