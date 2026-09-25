package updatecli

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
"updatecli" defines the specification for the Updatecli autodiscovery crawler.
It searches Updatecli compose files and generates manifests to update the policy versions they reference.
*/
type Spec struct {
	// "rootdir" defines the directory where the crawler starts searching for Updatecli compose files.
	//
	// default:
	//   the scm directory when "scmid" is set, otherwise the directory relative paths resolve from, by default the working directory.
	//
	// remark:
	//   * a relative path is resolved from the default directory.
	//   * an absolute path is used as is, instead of the scm directory.
	//
	RootDir string `yaml:",omitempty"`
	// "ignore" defines rules to exclude matching policies from the autodiscovery.
	//
	// remark:
	//   * a policy is ignored when it matches at least one rule.
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// "only" defines rules to restrict the autodiscovery to matching policies.
	//
	// remark:
	//   * a policy is kept only when it matches at least one rule.
	//
	Only MatchingRules `yaml:",omitempty"`
	// "files" defines the file name patterns the crawler searches for.
	//
	// The pattern syntax is:
	//
	// ```
	//     pattern:
	//         { term }
	//     term:
	//         '*'         matches any sequence of non-Separator characters
	//         '?'         matches any single non-Separator character
	//         '[' [ '^' ] { character-range } ']'
	//                     character class (must be non-empty)
	//         c           matches character c (c != '*', '?', '\\', '[')
	//         '\\' c      matches character c
	//
	//     character-range:
	//         c           matches character c (c != '\\', '-', ']')
	//         '\\' c      matches character c
	//         lo '-' hi   matches character c for lo <= c <= hi
	// ```
	//
	// default:
	//   ```
	//   - "update-compose.yaml"
	//   - "updatecli-compose.yaml"
	//   ```
	//
	// remark:
	//   * the pattern is matched against the file name only, not against its path.
	//   * the pattern must match the whole file name, not just a substring.
	//   * on Windows, escaping is disabled and `\\` is treated as a path separator.
	//
	Files []string `yaml:",omitempty"`
	// "auths" defines the registry credentials, keyed by registry host without scheme.
	//
	// remark:
	//   * not passed to the generated manifests yet, they use the local OCI credentials, such as the Docker ones.
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

// Updatecli hold all information needed to generate updatecli manifest.
type Updatecli struct {
	// actionID holds the actionID used by the newly generated manifest
	actionID string
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootdir defines the root directory from where looking for Updatecli
	rootDir string
	// scmID holds the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	// files holds the list of files to analyze
	files []string
}

// New return a new valid Updatecli object.
func New(spec interface{}, rootDir, scmID, actionID string) (Updatecli, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Updatecli{}, err
	}

	// Validate ignore rules
	if err := s.Ignore.Validate(); err != nil {
		return Updatecli{}, fmt.Errorf("invalid ignore spec: %w", err)
	}

	// Validate only rules
	if err := s.Only.Validate(); err != nil {
		return Updatecli{}, fmt.Errorf("invalid only spec: %w", err)
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
		return Updatecli{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, Updatecli policies versioning use semantic versioning
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	files := DefaultFiles
	if len(s.Files) > 0 {
		files = s.Files
	}

	return Updatecli{
		actionID:      actionID,
		files:         files,
		spec:          s,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
	}, nil

}

// DiscoverManifests search for Updatecli compose file and generate Updatecli manifests.
func (u Updatecli) DiscoverManifests() ([][]byte, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Updatecli"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Updatecli")+1))

	return u.discoverUpdatecliPolicyManifests()
}
