package flux

import (
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
	defaultVersionFilterPattern string   = "*"
	defaultVersionFilterKind    string   = "semver"
)

// Spec defines the parameters which can be provided to the Flux crawler.
type Spec struct {
	// auths provides a map of registry credentials where the key is the registry URL without scheme
	Auths map[string]docker.InlineKeyChain `yaml:",omitempty"`
	// digest allows to specify if the generated manifest should use OCI digest on top of the tag
	//
	// default: true
	Digest *bool `yaml:",omitempty"`
	// helmRelease define if helmrelease file should be updated or not
	//
	// default: true
	HelmRelease *bool `yaml:",omitempty"`
	// files allows to override default flux files
	//
	// default: ["*.yaml", "*.yml"]
	Files []string `yaml:",omitempty"`
	// ignore allows to specify rule to ignore autodiscovery a specific Flux helmrelease based on a rule
	//
	// default: empty
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// only allows to specify rule to only autodiscover manifest for a specific Flux helm release based on a rule
	//
	// default: empty
	//
	Only MatchingRules `yaml:",omitempty"`
	// OCIRepository allows to specify if OCI repository files should be updated
	//
	// default: true
	OCIRepository *bool `yaml:",omitempty"`
	// rootDir defines the root directory used to recursively search for Flux files
	//
	// default: . (current working directory) or scm root directory
	//
	RootDir string `yaml:",omitempty"`
	// versionfilter provides parameters to specify the version pattern used when generating manifest.
	//
	// kind - semver
	//		versionfilter of kind `semver` uses semantic versioning as version filtering
	//		pattern accepts one of:
	//			`patch` - patch only update patch version
	//			`minor` - minor only update minor version
	//			`major` - major only update major versions
	//			`a version constraint` such as `>= 1.0.0`
	//
	//	kind - regex
	//		versionfilter of kind `regex` uses regular expression as version filtering
	//		pattern accepts a valid regular expression
	//
	//	example:
	//	```
	//		versionfilter:
	//			kind: semver
	//			pattern: minor
	//	```
	//
	//	and its type like regex, semver, or just latest.
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
