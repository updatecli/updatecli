package dockerfile

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
"dockerfile" defines the specification for the Dockerfile autodiscovery crawler.
It searches Dockerfiles and generates manifests to update the container images used by their "FROM" instructions.
*/
type Spec struct {
	// "digest" defines whether the generated manifests pin the image digest in addition to the tag.
	//
	// default:
	//   true
	//
	Digest *bool `yaml:",omitempty"`
	// "rootdir" defines the directory where the crawler starts searching for Dockerfiles.
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
	// "filematch" defines the file name patterns used to identify Dockerfiles.
	//
	// default:
	//   ```
	//   filematch:
	//     - "Dockerfile"
	//     - "Dockerfile.*"
	//   ```
	//
	// remark:
	//   * the pattern is matched against the file name only, not against its path.
	//   * the pattern follows the Go filepath.Match syntax, such as "*" or "?".
	//
	FileMatch []string `yaml:",omitempty"`
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

// Dockerfile hold all information needed to generate Dockerfile manifest.
type Dockerfile struct {
	// digest holds the value of the digest parameter
	digest bool
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for Helm Chart
	rootDir string
	// filematch defines the filematch rule used to identify the Dockerfile that need to be handled
	filematch []string
	// actionID hold the actionID used by the newly generated manifest
	actionID string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid Helm object.
func New(spec interface{}, rootDir, scmID, actionID string) (Dockerfile, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Dockerfile{}, err
	}

	// Validate ignore rules
	if err := s.Ignore.Validate(); err != nil {
		return Dockerfile{}, fmt.Errorf("invalid ignore spec: %w", err)
	}

	// Validate only rules
	if err := s.Only.Validate(); err != nil {
		return Dockerfile{}, fmt.Errorf("invalid only spec: %w", err)
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
		return Dockerfile{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, helm versioning uses semantic versioning. Containers is not but...
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	digest := true
	if s.Digest != nil {
		digest = *s.Digest
	}

	d := Dockerfile{
		actionID:      actionID,
		digest:        digest,
		spec:          s,
		rootDir:       dir,
		filematch:     DefaultFileMatch,
		scmID:         scmID,
		versionFilter: newFilter,
	}

	if len(s.FileMatch) > 0 {
		d.filematch = s.FileMatch
	}

	return d, nil

}

func (d Dockerfile) DiscoverManifests() ([][]byte, error) {

	logrus.Infof("\n\n%s\n", strings.ToTitle("Dockerfile"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Dockerfile")+1))

	return d.discoverDockerfileManifests()
}
