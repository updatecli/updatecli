// Package pyproject implements the autodiscovery crawler for Python projects.
//
// It walks a root directory looking for pyproject.toml files and generates one manifest per
// dependency declared in [project.dependencies] and [project.optional-dependencies]. Other
// tables, such as [dependency-groups], [tool.poetry], and [build-system], are not read.
//
// The package manager is detected from the lock file sitting next to each pyproject.toml.
// Only uv is supported today, through uv.lock:
//
//   - uv.lock present and the uv command available: a pypi source and a shell target running
//     `uv lock --upgrade-package` are generated. Only uv.lock is rewritten; the constraints
//     declared in pyproject.toml are left untouched.
//   - uv.lock present but the uv command missing: the whole pyproject.toml is skipped, since
//     Updatecli cannot re-lock what it would bump.
//   - no lock file: only the pypi source is generated, so the latest version is reported but
//     nothing is modified.
//
// Dependency strings are parsed as PEP 508 specifiers. Environment markers are stripped rather
// than evaluated, extras are dropped from the tracked package name, and direct references such
// as `mypkg @ https://...` are skipped.
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
	//  `versionfilter` provides parameters to specify the version pattern used when generating manifest.
	//
	//  If unspecified, Updatecli falls back to kind `pep440` and reuses each dependency's own
	//  constraint as the pattern, such as `>=2.28` for `requests>=2.28`, or `*` when the
	//  dependency is declared without a constraint.
	//
	//  kind - pep440 (default)
	//    versionfilter of kind `pep440` uses PEP 440 version specifiers natively
	//    pattern accepts a PEP 440 version specifier such as `>=2.28`, `>=1.0,<3.0`, or `*` (any)
	//
	//  kind - semver
	//    versionfilter of kind `semver` uses semantic versioning as version filtering
	//    pattern accepts one of:
	//      `prerelease` - Updatecli tries to identify the latest prerelease whatever it means
	//      `patch` - Updatecli only handles patch version update
	//      `minor` - Updatecli handles patch AND minor version update
	//      `minoronly` - Updatecli handles minor version only
	//      `major` - Updatecli handles patch, minor, AND major version update
	//      `majoronly` - Updatecli only handles major version update
	//      `a version constraint` such as `>= 1.0.0`
	//    relative patterns such as `minor` are resolved against the version currently declared
	//    by each dependency, so `minor` generates the pattern `2.x` for `requests>=2.28`
	//
	//  kind - regex
	//    versionfilter of kind `regex` uses regular expression as version filtering
	//    pattern accepts a valid regular expression
	//
	//  example:
	//  ```
	//    versionfilter:
	//      kind: pep440
	//      pattern: ">=2.28"
	//  ```
	//
	//  More examples can be found at https://www.updatecli.io/docs/core/versionfilter/
	VersionFilter version.Filter `yaml:",omitempty"`
	// IndexURL specifies a custom PyPI index URL propagated to all generated source specs.
	// It carries no credentials: authenticating against a private registry requires setting the
	// pypi resource `token` field on the generated manifests.
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
		logrus.Debugln("no versioning filter specified, falling back to pep440 versioning")
		newFilter.Kind = version.PEP440VERSIONKIND
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
