package pyproject

import (
	"fmt"
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Spec defines the pyproject autodiscovery parameters.
type Spec struct {
	// RootDir defines the root directory used to recursively search for pyproject.toml files.
	RootDir string `yaml:",omitempty"`
	// Ignore specifies rules to exclude pyproject.toml dependencies from autodiscovery.
	Ignore MatchingRules `yaml:",omitempty"`
	// Only specifies rules to restrict autodiscovery to matching pyproject.toml dependencies.
	Only MatchingRules `yaml:",omitempty"`
	/*
		versionfilter provides parameters to specify the version pattern used when generating manifests.

		kind - semver
			versionfilter of kind `semver` uses semantic versioning as version filtering.
			pattern accepts one of:
				`patch` - update patch version only
				`minor` - update minor version only
				`major` - update major versions only
				`a version constraint` such as `>= 1.0.0`

		kind - regex
			versionfilter of kind `regex` uses regular expression as version filtering.
			pattern accepts a valid regular expression.

		example:
		```
			versionfilter:
				kind: semver
				pattern: minor
		```
	*/
	VersionFilter version.Filter `yaml:",omitempty"`
	// IndexURL specifies a custom PyPI index URL propagated to all generated source specs.
	IndexURL string `yaml:",omitempty"`
}

// Pyproject holds all state needed to discover pyproject.toml dependency manifests.
type Pyproject struct {
	// spec is the user-supplied configuration.
	spec Spec
	// rootDir is the resolved directory to search from.
	rootDir string
	// actionID is propagated to generated manifests.
	actionID string
	// scmID is propagated to generated manifests.
	scmID string
	// versionFilter is the resolved filter (may differ from spec.VersionFilter when defaults apply).
	versionFilter version.Filter
	// uvAvailable reports whether the uv CLI is present on PATH.
	uvAvailable bool
}

// New constructs a valid Pyproject autodiscovery instance from the provided spec.
func New(spec interface{}, rootDir, scmID, actionID string) (Pyproject, error) {
	var s Spec

	if err := mapstructure.Decode(spec, &s); err != nil {
		return Pyproject{}, err
	}

	if err := s.Ignore.Validate(); err != nil {
		return Pyproject{}, fmt.Errorf("invalid ignore spec: %w", err)
	}

	if err := s.Only.Validate(); err != nil {
		return Pyproject{}, fmt.Errorf("invalid only spec: %w", err)
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
		return Pyproject{}, fmt.Errorf("no working directory defined")
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		logrus.Debugln("no versioning filter specified, falling back to semantic versioning")
		newFilter.Kind = version.SEMVERVERSIONKIND
		newFilter.Pattern = "*"
	}

	return Pyproject{
		actionID:      actionID,
		spec:          s,
		rootDir:       dir,
		scmID:         scmID,
		versionFilter: newFilter,
		uvAvailable:   isUvAvailable(),
	}, nil
}

// DiscoverManifests returns updatecli manifests for all Python dependencies found under rootDir.
func (p Pyproject) DiscoverManifests() ([][]byte, error) {
	logrus.Infof("\n\n%s\n", strings.ToTitle("Pyproject"))
	logrus.Infof("%s\n", strings.Repeat("=", len("Pyproject")+1))

	return p.discoverDependencyManifests()
}
