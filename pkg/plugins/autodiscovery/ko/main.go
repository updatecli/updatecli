package ko

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
"ko" defines the specification for the Ko autodiscovery crawler.
It searches Ko configuration files and generates manifests to update the base container images they use.
*/
type Spec struct {
	// "auths" defines the registry credentials, keyed by registry host without scheme.
	//
	// remark:
	//   * when empty, Updatecli uses the local OCI credentials, such as the Docker ones.
	//
	// example:
	//   ```
	//   auths:
	//     "ghcr.io":
	//       token: "xxx"
	//     "index.docker.io":
	//       username: "admin"
	//       password: "password"
	//   ```
	//
	Auths map[string]docker.InlineKeyChain `yaml:",omitempty"`
	// "digest" defines whether the generated manifests pin the image digest in addition to the tag.
	//
	// default:
	//   true
	//
	Digest *bool `yaml:",omitempty"`
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
	//   - ".ko.yaml"
	//   ```
	//
	// remark:
	//   * the pattern is matched against the file name only, not against its path.
	//   * the pattern must match the whole name, not just a substring.
	//   * on Windows, escaping is disabled and `\\` is treated as a path separator.
	//
	Files []string `yaml:",omitempty"`
	// "rootdir" defines the directory where the crawler starts searching for Ko files.
	//
	// default:
	//   the scm directory when "scmid" is set, otherwise the directory relative paths resolve from, by default the working directory.
	//
	// remark:
	//   * a relative path is resolved from the default directory.
	//   * an absolute path is used as is, instead of the scm directory.
	//
	RootDir string `yaml:",omitempty"`
	// "ignore" defines rules to exclude matching container images from the autodiscovery.
	//
	// remark:
	//   * a container image is ignored when it matches at least one rule.
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// "only" defines rules to restrict the autodiscovery to matching container images.
	//
	// remark:
	//   * a container image is kept only when it matches at least one rule.
	//
	Only MatchingRules `yaml:",omitempty"`
	// "versionfilter" defines the version filter used by the generated manifests.
	//
	// default:
	//   kind "semver" with pattern ">=<current tag>", combined with a tag filter derived from the current tag.
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

// Ko holds all information needed to generate Ko manifests.
type Ko struct {
	// actionID holds the actionID used by the newly generated manifest
	actionID string
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for Kubernetes manifests
	rootDir string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	// files holds the list of files to analyze
	files []string
	// digest holds the value of the digest parameter
	digest bool
}

// New return a new valid Ko object.
func New(spec interface{}, rootDir, scmID, actionID string) (Ko, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Ko{}, err
	}

	// Validate ignore rules
	if err := s.Ignore.Validate(); err != nil {
		return Ko{}, fmt.Errorf("invalid ignore spec: %w", err)
	}

	// Validate only rules
	if err := s.Only.Validate(); err != nil {
		return Ko{}, fmt.Errorf("invalid only spec: %w", err)
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
		return Ko{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, helm versioning uses semantic versioning. Containers is not but...
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	files := DefaultKoFiles
	if len(s.Files) > 0 {
		files = s.Files
	}

	digest := true
	if s.Digest != nil {
		digest = *s.Digest
	}

	return Ko{
		actionID:      actionID,
		digest:        digest,
		spec:          s,
		rootDir:       dir,
		files:         files,
		scmID:         scmID,
		versionFilter: newFilter,
	}, nil

}

func (f Ko) DiscoverManifests() ([][]byte, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Ko"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Ko")+1))

	return f.discoverContainerManifests()
}
