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

/*
"pyproject" defines the specification for the pyproject autodiscovery crawler.
It searches "pyproject.toml" files and generates manifests to update the Python dependencies they declare.
*/
type Spec struct {
	// "rootdir" defines the directory where the crawler starts searching for "pyproject.toml" files.
	//
	// default:
	//   the scm directory when "scmid" is set, otherwise the directory relative paths resolve from, by default the working directory.
	//
	// remark:
	//   * a relative path is resolved from the default directory.
	//   * an absolute path is used as is, instead of the scm directory.
	//
	RootDir string `yaml:",omitempty"`
	// "ignore" defines rules to exclude matching Python dependencies from the autodiscovery.
	//
	// remark:
	//   * a Python dependency is ignored when it matches at least one rule.
	//
	Ignore MatchingRules `yaml:",omitempty"`
	// "only" defines rules to restrict the autodiscovery to matching Python dependencies.
	//
	// remark:
	//   * a Python dependency is kept only when it matches at least one rule.
	//
	Only MatchingRules `yaml:",omitempty"`
	// "versionfilter" defines the version filter used by the generated manifests.
	//
	// default:
	//   kind "pep440" with the dependency's own constraint as pattern, such as ">=2.28" for
	//   "requests>=2.28", or "*" when the dependency is declared without a constraint.
	//
	// remark:
	//   * with kind "pep440", "pattern" accepts a PEP 440 version specifier, such as ">=2.28", ">=1.0,<3.0" or "*".
	//   * with kind "semver", "pattern" accepts:
	//     * "prerelease": the latest prerelease of the current version.
	//     * "patch": patch updates only.
	//     * "minor": patch and minor updates.
	//     * "minoronly": minor updates only.
	//     * "major": patch, minor and major updates.
	//     * "majoronly": major updates only.
	//     * a version constraint, such as ">= 1.0.0".
	//   * relative patterns such as "minor" are resolved against the version each dependency
	//     currently declares, so "minor" gives the pattern "2.x" for "requests>=2.28".
	//   * with kind "regex", "pattern" accepts a regular expression.
	//   * more examples at https://www.updatecli.io/docs/core/versionfilter/
	//
	// example:
	//   ```
	//   versionfilter:
	//     kind: pep440
	//     pattern: ">=2.28"
	//   ```
	//
	VersionFilter version.Filter `yaml:",omitempty"`
	// "indexurl" defines a custom PyPI index URL used by every generated source.
	//
	// remark:
	//   * it carries no credentials: a private registry requires setting the pypi
	//     resource "token" field in the generated manifests.
	//
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
