package flux

import (
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/docker"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

var (
	// defaultFluxFiles specifies accepted Helm chart metadata file name
	defaultFluxFiles            []string = []string{"*.yaml", "*.yml"}
	defaultVersionFilterPattern string   = "*"
	defaultVersionFilterKind    string   = "semver"
)

// Spec defines the parameters which can be provided to the Flux crawler.
type Spec struct {
	/*
		auths provides a map of registry credentials where the key is the registry URL without scheme
	*/
	Auths map[string]docker.InlineKeyChain `yaml:",omitempty"`
	/*
		Digest allows to specify if the generated manifest should use OCI digest on top of the tag
	*/
	Digest *bool `yaml:",omitempty"`

	/*
	   fileMatch allows to override default flux files
	*/
	Files []string `yaml:",omitempty"`
	/*
	   Ignore allows to specify rule to ignore autodiscovery a specific Flux helmrelease based on a rule

	   default: empty
	*/
	Ignore MatchingRules `yaml:",omitempty"`
	/*
		Only allows to specify rule to only autodiscover manifest for a specific Flux helm release based on a rule

		default: empty
	*/
	Only MatchingRules `yaml:",omitempty"`
	/*
	   RootDir defines the root directory used to recursively search for Flux files

	   default: . (current working directory) or scm root directory
	*/
	RootDir string `yaml:",omitempty"`
	/*
		versionfilter provides parameters to specify the version pattern used when generating manifest.

		kind - semver
			versionfilter of kind `semver` uses semantic versioning as version filtering
			pattern accepts one of:
				`patch` - patch only update patch version
				`minor` - minor only update minor version
				`major` - major only update major versions
				`a version constraint` such as `>= 1.0.0`

		kind - regex
			versionfilter of kind `regex` uses regular expression as version filtering
			pattern accepts a valid regular expression

		example:
		```
			versionfilter:
				kind: semver
				pattern: minor
		```

		and its type like regex, semver, or just latest.
	*/
	VersionFilter version.Filter `yaml:",omitempty"`
}

// Flux holds all information needed to generate Flux manifest.
type Flux struct {
	// files defines the accepted Flux file name
	files []string
	// digest defines if the generated manifest should use OCI digest on top of the tag
	digest bool
	// spec defines the settings provided via an updatecli manifest
	spec Spec
	// rootDir defines the  oot directory from where looking for Flux
	rootDir string
	// scmID hold the scmID used by the newly generated manifest
	scmID string
	// versionFilter holds the "valid" version.filter, that might be different from the user-specified filter (Spec.VersionFilter)
	versionFilter version.Filter
	// helmRepositories is a list of HelmRepository
	helmRepositories []helmRepository
	// ociRepositories is a list of OCIRepository files found
	ociRepositoryFiles []string
	// helmReleaseFiles is a list of HelmRelease files found
	helmReleaseFiles []string
}

// New return a new valid Flux object.
func New(spec interface{}, rootDir, scmID string) (Flux, error) {
	var s Spec

	err := mapstructure.Decode(spec, &s)
	if err != nil {
		return Flux{}, err
	}

	dir := rootDir
	if len(s.RootDir) > 0 {
		dir = s.RootDir
	}

	// If no RootDir have been provided via settings,
	// then fallback to the current process path.
	if len(dir) == 0 {
		logrus.Errorln("no working directory defined")
		return Flux{}, err
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
		digest:        digest,
		spec:          s,
		files:         files,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
	}, nil

}

func (f Flux) DiscoverManifests() ([][]byte, error) {
	var manifests [][]byte

	logrus.Infof("\n\n%s\n", strings.ToTitle("Flux"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Flux")+1))

	err := f.searchFluxFiles(f.rootDir, f.files)
	if err != nil {
		return nil, err
	}

	helmReleasemanifests := f.discoverHelmreleaseManifests()
	ociRepositoryManifests := f.discoverOCIRepositoryManifests()

	manifests = append(manifests, helmReleasemanifests...)
	manifests = append(manifests, ociRepositoryManifests...)

	return manifests, err
}
