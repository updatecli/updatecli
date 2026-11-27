package woodpecker

import (
	"fmt"
	"path"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"woodpecker" defines the specification for the Woodpecker autodiscovery crawler.
It searches Woodpecker workflow files and generates manifests to update the container images they use.
*/
type Spec struct {
	// "digest" defines whether the generated manifests pin the image digest in addition to the tag.
	//
	// default:
	//   true
	//
	Digest *bool `yaml:",omitempty"`
	// "rootdir" defines the directory where the crawler starts searching for Woodpecker workflow files.
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
	// "registrymappings" maps an image name prefix, as written in the Woodpecker workflow files,
	// to the image name prefix where Updatecli looks up new tags and digests.
	//
	// default:
	//   empty, new versions are looked up under the image name written in the files.
	//
	// remark:
	//   * use it for images pulled through a mirror, so new versions are looked up upstream
	//     while the files keep the mirror name.
	//   * a prefix matches whole path segments only: "my-registry.com/docker-io" matches
	//     "my-registry.com/docker-io/nginx" but not "my-registry.com/docker-iox/nginx".
	//   * when several prefixes match, the longest one wins.
	//   * a prefix can be a full image name, to map a single image, for instance
	//     when one mirror holds images from several upstream registries.
	//   * "auths" credentials are picked for the upstream registry for the sources,
	//     and for the mirror registry for the condition.
	//   * "ignore" and "only" rules match the image name written in the files.
	//   * a condition checks the mirror already serves the new tag, or the new tag
	//     and digest with "digest" enabled, so the files are only updated once the
	//     mirror has caught up. With "digest" enabled, the mirror must serve the
	//     image unchanged for its digest to match the upstream one.
	//
	// example:
	//   ```
	//   registrymappings:
	//     "my-registry.com/docker-io/": "docker.io/"
	//     "my-registry.com/ghcr-io/": "ghcr.io/"
	//     # a single image mirrored under another name
	//     "my-registry.com/tools/updatecli": "ghcr.io/updatecli/updatecli"
	//   ```
	//
	RegistryMappings docker.RegistryMappings `yaml:",omitempty"`
	// "filematch" defines the file path patterns the crawler searches for.
	//
	// default:
	//   ```
	//   - ".woodpecker.yml"
	//   - ".woodpecker.yaml"
	//   - ".woodpecker/*.yml"
	//   - ".woodpecker/*.yaml"
	//   - ".woodpecker/**/*.yml"
	//   - ".woodpecker/**/*.yaml"
	//   ```
	//
	// remark:
	//   * a pattern is matched against the file path relative to "rootdir", then against the file name.
	//   * the pattern follows the Go filepath.Match syntax, such as "*" or "?".
	//   * "**" is not recursive. It behaves like "*" and matches a single directory level,
	//     so the default patterns find files directly under ".woodpecker" or one directory below it.
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

// Woodpecker holds all information needed to generate Woodpecker workflow manifest.
type Woodpecker struct {
	// digest holds the value of the digest parameter
	digest bool
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the root directory from where looking for Woodpecker workflows
	rootDir string
	// filematch defines the filematch rule used to identify Woodpecker workflows that need to be handled
	filematch []string
	// actionID holds the actionID used by the newly generated manifest
	actionID string
	// scmID holds the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid Woodpecker object.
func New(spec interface{}, rootDir, scmID, actionID string) (Woodpecker, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Woodpecker{}, err
	}

	// Validate ignore rules
	if err := s.Ignore.Validate(); err != nil {
		return Woodpecker{}, fmt.Errorf("invalid ignore spec: %w", err)
	}

	// Validate only rules
	if err := s.Only.Validate(); err != nil {
		return Woodpecker{}, fmt.Errorf("invalid only spec: %w", err)
	}

	if err := s.RegistryMappings.Validate(); err != nil {
		return Woodpecker{}, err
	}

	dir := rootDir
	if path.IsAbs(s.RootDir) {
		if scmID != "" {
			log.Warningf("rootdir %q is an absolute path, scmID %q will be ignored", s.RootDir, scmID)
		}
		dir = s.RootDir
	}

	// If no RootDir have been provided via settings,
	// then fallback to the current process path.
	if len(dir) == 0 {
		log.Errorln("no working directory defined")
		return Woodpecker{}, err
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, use semantic versioning
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	digest := true
	if s.Digest != nil {
		digest = *s.Digest
	}

	w := Woodpecker{
		actionID:      actionID,
		digest:        digest,
		spec:          s,
		rootDir:       dir,
		filematch:     DefaultFilePatterns,
		scmID:         scmID,
		versionFilter: newFilter,
	}

	if len(s.FileMatch) > 0 {
		w.filematch = s.FileMatch
	}

	return w, nil
}

func (w Woodpecker) DiscoverManifests() ([][]byte, error) {
	log.Heading("Woodpecker")

	return w.discoverWorkflowImageManifests()
}
