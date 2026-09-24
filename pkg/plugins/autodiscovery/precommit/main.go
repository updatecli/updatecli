package precommit

import (
	"fmt"
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"precommit" defines the specification for the pre-commit autodiscovery crawler.
It searches ".pre-commit-config.yaml" files and generates manifests to update the hook repository revisions.
*/
type Spec struct {
	// "rootdir" defines the directory where the crawler starts searching for ".pre-commit-config.yaml" files.
	//
	// default:
	//   the scm directory when "scmid" is set, otherwise the directory relative paths resolve from, by default the working directory.
	//
	// remark:
	//   * a relative path is resolved from the default directory.
	//   * an absolute path is used as is, instead of the scm directory.
	//
	RootDir string `yaml:",omitempty"`
	// "ignore" defines rules to exclude matching hook repositories from the autodiscovery.
	//
	// remark:
	//   * a hook repository is ignored when it matches at least one rule.
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// "only" defines rules to restrict the autodiscovery to matching hook repositories.
	//
	// remark:
	//   * a hook repository is kept only when it matches at least one rule.
	//
	Only MatchingRules `yaml:",omitempty"`
	// "versionfilter" defines the version filter used by the generated manifests.
	//
	// default:
	//   kind "semver" with pattern "*", any version greater than or equal to the current one.
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
	//   * when "rev" is not a version, such as a commit hash, the version is read from the comment next to it.
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
	// "digest" defines whether the generated manifests pin the commit hash instead of the tag.
	//
	// default:
	//   false
	//
	// remark:
	//   * it matches the "--freeze" option of "pre-commit autoupdate", see https://pre-commit.com/#pre-commit-autoupdate
	//
	Digest *bool `yaml:",omitempty"`
}

// Precommit holds all information needed to generate precommit manifest.
type Precommit struct {
	// actionID holds the actionID used by the newly generated manifest
	actionID string
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for precommit
	rootDir string
	// scmID holds the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	// digest holds the value of the digest parameter
	digest bool
}

// New return a new valid object.
func New(spec interface{}, rootDir, scmID, actionID string) (Precommit, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Precommit{}, err
	}

	// Validate ignore rules
	if err := s.Ignore.Validate(); err != nil {
		return Precommit{}, fmt.Errorf("invalid ignore spec: %w", err)
	}

	// Validate only rules
	if err := s.Only.Validate(); err != nil {
		return Precommit{}, fmt.Errorf("invalid only spec: %w", err)
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
		return Precommit{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, Updatecli policies versioning use semantic versioning
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	digest := false
	if s.Digest != nil {
		digest = *s.Digest
	}

	return Precommit{
		actionID:      actionID,
		spec:          s,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
		digest:        digest,
	}, nil

}

func (p Precommit) DiscoverManifests() ([][]byte, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("PRECOMMIT"))
	logrus.Infof("%s\n", strings.Repeat("=", len("PRECOMMIT")+1))

	manifests, err := p.discoverDependencyManifests()

	if err != nil {
		return nil, err
	}

	return manifests, nil
}
