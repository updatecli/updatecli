package flux

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"

	fluxcd "github.com/fluxcd/source-controller/api/v1beta2"
)

var (
	// defaultFluxFiles specifies accepted Helm chart metadata file name
	defaultFluxFiles            []string = []string{"*.yaml", "*.yml"}
	defaultVersionFilterRegex   string   = "*"
	defaultVersionFilterPattern string   = "*"
	defaultVersionFilterKind    string   = "semver"
)

/*
"flux" defines the specification for the Flux autodiscovery crawler.
It searches Flux manifests and generates manifests to update the Helm charts of HelmRelease resources and the artifacts of OCIRepository resources.
*/
type Spec struct {
	// "auths" defines the registry credentials, keyed by registry host without scheme.
	//
	// remark:
	//   * when empty, Updatecli uses the local OCI credentials, such as the Docker ones.
	//   * for a HelmRelease chart, only the "token" is used, looked up by the Helm repository host.
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
	// "digest" defines whether the generated manifests pin the OCIRepository artifact digest in addition to the tag.
	//
	// default:
	//   true
	//
	Digest *bool `yaml:",omitempty"`
	// "helmrelease" defines whether HelmRelease resources are updated.
	//
	// default:
	//   true
	//
	HelmRelease *bool `yaml:",omitempty"`
	// "files" defines the file name patterns the crawler searches for.
	//
	// default:
	//   ```
	//   files:
	//     - "*.yaml"
	//     - "*.yml"
	//   ```
	//
	// remark:
	//   * the pattern is matched against the file name only, not against its path.
	//   * the pattern follows the Go filepath.Match syntax, such as "*" or "?".
	//
	Files []string `yaml:",omitempty"`
	// "ignore" defines rules to exclude matching artifacts from the autodiscovery.
	//
	// remark:
	//   * a artifact is ignored when it matches at least one rule.
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// "only" defines rules to restrict the autodiscovery to matching artifacts.
	//
	// remark:
	//   * a artifact is kept only when it matches at least one rule.
	//
	Only MatchingRules `yaml:",omitempty"`
	// "ocirepository" defines whether OCIRepository resources are updated.
	//
	// default:
	//   true
	//
	OCIRepository *bool `yaml:",omitempty"`
	// "rootdir" defines the directory where the crawler starts searching for Flux manifests.
	//
	// default:
	//   the scm directory when "scmid" is set, otherwise the directory relative paths resolve from, by default the working directory.
	//
	// remark:
	//   * a relative path is resolved from the default directory.
	//   * an absolute path is used as is, instead of the scm directory.
	//
	RootDir string `yaml:",omitempty"`
	// "versionfilter" defines the version filter used by the generated manifests.
	//
	// default:
	//   * for a HelmRelease chart, kind "semver" with pattern "*", the latest version.
	//   * for an OCIRepository artifact, kind "semver" with pattern ">=<current tag>", combined with a tag filter derived from the current tag.
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

// Flux holds all information needed to generate Flux manifest.
type Flux struct {
	// files defines the accepted Flux file name
	files []string
	// helmRelease defines if the generated manifest should be a HelmRelease
	helmRelease bool
	// digest defines if the generated manifest should use OCI digest on top of the tag
	digest bool
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the  oot directory from where looking for Flux
	rootDir string
	// actionID hold the actionID used by the newly generated manifest
	actionID string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
	// ociRepository defines if the OCI repository should be updated
	ociRepository bool
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	// helmRepositories is a list of HelmRepository
	helmRepositories []fluxcd.HelmRepository
	// ociRepositories is a list of OCIRepository files found
	ociRepositoryFiles []string
	// helmReleaseFiles is a list of HelmRelease files found
	helmReleaseFiles []string
}

// New return a new valid Flux object.
func New(spec interface{}, rootDir, scmID, actionID string) (Flux, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Flux{}, err
	}

	// Validate ignore rules
	if err := s.Ignore.Validate(); err != nil {
		return Flux{}, fmt.Errorf("invalid ignore spec: %w", err)
	}

	// Validate only rules
	if err := s.Only.Validate(); err != nil {
		return Flux{}, fmt.Errorf("invalid only spec: %w", err)
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
		return Flux{}, err
	}

	ociRepository := true
	if s.OCIRepository != nil {
		ociRepository = *s.OCIRepository
	}

	helmRelease := true
	if s.HelmRelease != nil {
		helmRelease = *s.HelmRelease
	}

	digest := true
	if s.Digest != nil {
		digest = *s.Digest
	}

	files := defaultFluxFiles
	if len(s.Files) > 0 {
		files = s.Files
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		// By default, helm versioning uses semantic versioning.
		newFilter.Kind = "semver"
		newFilter.Pattern = "*"
	}

	return Flux{
		actionID:      actionID,
		digest:        digest,
		spec:          s,
		files:         files,
		ociRepository: ociRepository,
		helmRelease:   helmRelease,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
	}, nil

}

func (f Flux) DiscoverManifests() ([][]byte, error) {
	var manifests [][]byte

	logrus.Infof("\n\n%s\n", strings.ToTitle("Flux"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Flux")+1))

	searchFromDir := f.rootDir
	// If the spec.RootDir is an absolute path, then it as already been set
	// correctly in the New function.
	if f.spec.RootDir != "" && !path.IsAbs(f.spec.RootDir) {
		searchFromDir = filepath.Join(f.rootDir, f.spec.RootDir)
	}

	err := f.searchFluxFiles(searchFromDir, f.files)
	if err != nil {
		return nil, err
	}

	if f.helmRelease {
		helmReleasemanifests := f.discoverHelmreleaseManifests()
		manifests = append(manifests, helmReleasemanifests...)
	}

	if f.ociRepository {
		ociRepositoryManifests := f.discoverOCIRepositoryManifests()
		manifests = append(manifests, ociRepositoryManifests...)
	}

	return manifests, err
}
