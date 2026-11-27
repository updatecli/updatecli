package helm

import (
	"fmt"
	"path"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

/*
"helm" defines the specification for the Helm autodiscovery crawler.
It searches Helm charts and generates manifests to update their chart dependencies and the container images set in their values files.
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
	// "registrymappings" maps an image name prefix, as written in the values files,
	// to the image name prefix where Updatecli looks up new tags and digests.
	//
	// default:
	//   empty, new versions are looked up under the image name written in the files.
	//
	// remark:
	//   * it applies to container images in values files only, not to chart dependencies.
	//   * when "registry" and "repository" are set as separate keys, a prefix matches
	//     the image name made of both, such as "my-registry.com/ghcr-io/updatecli/updatecli".
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
	// "digest" defines whether the generated manifests pin the image digest in addition to the tag.
	//
	// default:
	//   true
	//
	// remark:
	//   * it applies to container images only.
	//
	Digest *bool `yaml:",omitempty"`
	// "ignorecontainer" disables the container image updates.
	//
	// default:
	//   false
	//
	IgnoreContainer bool `yaml:",omitempty"`
	// "ignorechartdependency" disables the chart dependency updates.
	//
	// default:
	//   false
	//
	IgnoreChartDependency bool `yaml:",omitempty"`
	// "ignore" defines rules to exclude matching chart dependencies or container images from the autodiscovery.
	//
	// remark:
	//   * a chart dependency or container image is ignored when it matches at least one rule.
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// "rootdir" defines the directory where the crawler starts searching for Helm charts.
	//
	// default:
	//   the scm directory when "scmid" is set, otherwise the directory relative paths resolve from, by default the working directory.
	//
	// remark:
	//   * a relative path is resolved from the default directory.
	//   * an absolute path is used as is, instead of the scm directory.
	//
	RootDir string `yaml:",omitempty"`
	// "only" defines rules to restrict the autodiscovery to matching chart dependencies or container images.
	//
	// remark:
	//   * a chart dependency or container image is kept only when it matches at least one rule.
	//
	Only MatchingRules `yaml:",omitempty"`
	// "versionfilter" defines the version filter used by the generated manifests.
	//
	// default:
	//   * chart dependencies: kind "semver" with pattern "*", the latest version.
	//   * container images: kind "semver" with pattern ">=<current tag>", combined with a tag filter derived from the current tag.
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
	// "skippackaging" sets "skippackaging" on the generated helm targets.
	//
	// default:
	//   false
	//
	// remark:
	//   * see the "skippackaging" field of the helm resource.
	//
	SkipPackaging bool `yaml:",omitempty"`
	// "versionincrement" sets "versionincrement" on the generated helm targets.
	//
	// It defines how the chart version is bumped when the chart changes.
	//
	// default:
	//   minor, the helm target default.
	//
	// remark:
	//   * accepted values are a comma separated list of "major", "minor" and "patch",
	//     or one of "auto" or "none" on its own.
	//
	// example:
	//   * versionincrement: patch
	//   * versionincrement: none
	//
	VersionIncrement string `yaml:",omitempty"`
}

// Helm hold all information needed to generate helm manifest.
type Helm struct {
	// digest holds the value of the digest parameter
	digest bool
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootdir defines the root directory from where looking for Helm Chart
	rootDir string
	// actionID hold the actionID used by the newly generated manifest
	actionID string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
}

// New return a new valid Helm object.
func New(spec interface{}, rootDir, scmID, actionID string) (Helm, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Helm{}, err
	}

	// Validate ignore rules
	if err := s.Ignore.Validate(); err != nil {
		return Helm{}, fmt.Errorf("invalid ignore spec: %w", err)
	}

	// Validate only rules
	if err := s.Only.Validate(); err != nil {
		return Helm{}, fmt.Errorf("invalid only spec: %w", err)
	}

	if err := s.RegistryMappings.Validate(); err != nil {
		return Helm{}, err
	}

	dir := rootDir
	if path.IsAbs(s.RootDir) {
		if scmID != "" {
			log.Warningf("rootdir %q is an absolute path, scmID %q will be ignored", s.RootDir, scmID)
		}
		dir = s.RootDir
	}

	// Fallback to the current process path if not rootdir specified.
	if len(dir) == 0 {
		log.Errorln("no working directory defined")
		return Helm{}, err
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

	return Helm{
		actionID:      actionID,
		digest:        digest,
		spec:          s,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
	}, nil

}

func (h Helm) DiscoverManifests() ([][]byte, error) {

	log.Heading("Helm")

	var manifests [][]byte
	var err error

	if !h.spec.IgnoreChartDependency {
		manifests, err = h.discoverHelmDependenciesManifests()
		if err != nil {
			log.Errorf("generating Helm chart dependencies manifest(s): %s", err)
		}
	}

	if !h.spec.IgnoreContainer {
		containerManifest, err := h.discoverHelmContainerManifests()
		if err != nil {
			log.Errorf("generating container update manifest(s): %s", err)
		}

		manifests = append(manifests, containerManifest...)
	}

	return manifests, nil
}
