// Package pyproject implements the autodiscovery crawler for Python projects.
//
// It walks a root directory looking for pyproject.toml files and generates one manifest per
// dependency declared in [project.dependencies], [project.optional-dependencies], and
// [dependency-groups] (PEP 735). In dependency groups, {include-group = "..."} entries don't
// generate manifests since the included group is processed on its own, but "dependencygroups"
// rules consider the dependencies of an included group part of the including one too.
// Other tables, such as [tool.poetry], [tool.uv] dev-dependencies, and [build-system], are
// not read.
//
// The package manager is detected from the lock file sitting next to each pyproject.toml.
// Only uv is supported today, through uv.lock:
//
//   - uv.lock present and the uv command available: a pypi source and a shell target running
//     `uv lock --upgrade-package` are generated. Only uv.lock is rewritten; the constraints
//     declared in pyproject.toml are left untouched.
//   - uv.lock present but the uv command missing: the whole pyproject.toml is skipped, since
//     Updatecli cannot re-lock what it would bump. With "vulnerability", uv.lock is still read
//     and the manifests only report the vulnerable packages.
//   - no lock file: only the pypi source is generated, so the latest version is reported but
//     nothing is modified.
//
// With "vulnerability" set, the crawler switches to security updates: a manifest is only
// generated for a dependency whose current version, from uv.lock or an exact pin, has known
// vulnerabilities in the OSV database, and it updates it to the lowest version without any.
//
// Generated sources carry an age filter, by default a minimum of seven days, so a release
// is only suggested once it has been published long enough to be yanked if it is broken
// or malicious.
//
// Dependency strings are parsed as PEP 508 specifiers. Environment markers are stripped rather
// than evaluated, extras are dropped from the tracked package name, and direct references such
// as `mypkg @ https://...` are skipped.
package pyproject

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/updatecli/updatecli/pkg/core/log"
	"github.com/updatecli/updatecli/pkg/plugins/utils/age"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
	"github.com/updatecli/updatecli/pkg/plugins/utils/vulnerability"
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
	// "age" defines the minimum or maximum age of a release to be considered valid.
	//
	// default:
	//   ```
	//   age:
	//     minimum: 7d
	//   ```
	//
	// remark:
	//   * it accepts a duration string, such as "24h", "7d", "3w" or "1y".
	//   * it is propagated to the generated pypi sources.
	//   * the default keeps Updatecli from suggesting a package version published less than
	//     seven days ago. An empty "age: {}" disables it.
	//   * the default doesn't apply with "indexurl", as private registries may not report the
	//     upload times an age filter needs. On a registry reporting none, an explicit "age"
	//     makes the generated sources fail.
	//   * security fixes don't wait for it, see "vulnerability".
	//
	Age *age.Spec `yaml:",omitempty"`
	// "vulnerability" switches the autodiscovery to security updates, based on the OSV database (https://osv.dev).
	//
	// A package is only updated when its current version has known vulnerabilities,
	// to the lowest version without any.
	//
	// remark:
	//   * only security updates are generated, routine updates require a separate manifest.
	//   * it cannot be combined with:
	//     * "age", as a fixed version is suggested as soon as it is known. The default age
	//       doesn't apply either.
	//     * "versionfilter", as the version is the lowest one without known vulnerabilities.
	//     * "indexurl", as the OSV database only covers packages published on pypi.org.
	//   * the current version is the version resolved in the uv.lock next to the pyproject.toml or,
	//     without one, the version pinned with "==" in pyproject.toml.
	//     Packages without an identifiable current version are ignored, as are packages uv.lock
	//     resolves from a workspace, a path or a git repository rather than from a registry.
	//   * packages uv.lock resolves from a registry other than pypi.org are skipped and reported,
	//     unless the registry is listed under "pypimirrors".
	//   * "only" and "ignore" version rules match the current version.
	//   * a package listed in several groups at the same version gives a single manifest.
	//   * pyproject.toml is never modified: when the fixed version falls outside the declared
	//     version constraint, the uv target fails until the constraint is widened.
	//   * without uv.lock, nothing is modified and the manifests only report the vulnerable packages.
	//   * labels, such as "security", are set on the action used by the manifest.
	//
	// example:
	//   ```
	//   vulnerability:
	//     minseverity: high
	//     ignore:
	//       - CVE-2024-56201
	//   ```
	//
	Vulnerability *vulnerability.Spec `yaml:",omitempty"`
	// "pypimirrors" defines the registries mirroring pypi.org, such as an Artifactory or Nexus
	// proxy, whose packages are checked against the OSV database by "vulnerability".
	//
	// default:
	//   empty, so only the packages uv.lock resolves from pypi.org are checked.
	//
	// remark:
	//   * it requires "vulnerability".
	//   * the OSV database only covers packages published on pypi.org. A package uv.lock resolves
	//     from another registry is skipped and reported unless its registry is listed here.
	//   * only list registries serving pypi.org packages: a private package may only share its
	//     name with a public one, and would get its advisories.
	//   * each value is the registry URL as written in uv.lock, a trailing "/" is ignored.
	//
	// example:
	//   ```
	//   pypimirrors:
	//     - https://artifactory.example.com/api/pypi/pypi-remote/simple
	//   ```
	//
	PyPIMirrors []string `yaml:",omitempty"`
}

// validateVulnerability checks the vulnerability spec and the settings it can't be combined with.
func (s Spec) validateVulnerability() error {
	if s.Vulnerability == nil {
		if len(s.PyPIMirrors) > 0 {
			return errors.New("pypimirrors requires vulnerability, it only selects the packages checked against the OSV database")
		}
		return nil
	}

	for i, mirror := range s.PyPIMirrors {
		u, err := url.Parse(strings.TrimSpace(mirror))
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return fmt.Errorf("pypimirrors entry %d %q is not an http or https URL", i+1, mirror)
		}
	}

	if err := s.Vulnerability.Validate(); err != nil {
		return fmt.Errorf("invalid vulnerability spec: %w", err)
	}

	if s.Age != nil && !s.Age.IsZero() {
		return errors.New("age can't be combined with vulnerability, use a separate manifest for routine updates")
	}

	if !s.VersionFilter.IsZero() {
		return errors.New("versionfilter can't be combined with vulnerability, the version is the lowest one without known vulnerabilities")
	}

	if s.IndexURL != "" {
		return errors.New("indexurl can't be combined with vulnerability, the OSV database only covers packages published on pypi.org")
	}

	return nil
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
	// releaseAge holds the release age filter to propagate to generated manifests.
	releaseAge age.Spec
	// trustedRegistries holds the normalized URLs of pypi.org and the mirrors from the spec,
	// whose packages vulnerability mode checks against the OSV database.
	trustedRegistries map[string]bool
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

	// Patterns are matched against normalized group names, so they are normalized once here.
	s.Ignore = s.Ignore.normalized()
	s.Only = s.Only.normalized()

	if err := s.validateVulnerability(); err != nil {
		return Pyproject{}, err
	}

	// By default, wait for a release to be old enough before suggesting it
	releaseAge := age.Spec{Minimum: defaultMinimumReleaseAge}
	switch {
	case s.Age != nil:
		releaseAge = *s.Age
	case s.Vulnerability != nil:
		// A fixed version is suggested as soon as it is known
		releaseAge = age.Spec{}
	case s.IndexURL != "":
		// Private registries may not report upload times, which an age filter needs
		releaseAge = age.Spec{}
	}

	if err := releaseAge.Validate(); err != nil {
		return Pyproject{}, fmt.Errorf("wrong age spec: %w", err)
	}

	dir := rootDir
	if path.IsAbs(s.RootDir) {
		if scmID != "" {
			log.Warningf("rootdir %q is an absolute path, scmID %q will be ignored", s.RootDir, scmID)
		}
		dir = s.RootDir
	}

	if len(dir) == 0 {
		log.Errorln("no working directory defined")
		return Pyproject{}, fmt.Errorf("no working directory defined")
	}

	newFilter := s.VersionFilter
	if s.VersionFilter.IsZero() {
		log.Debugln("no versioning filter specified, falling back to pep440 versioning")
		newFilter.Kind = version.PEP440VERSIONKIND
		newFilter.Pattern = "*"
	}

	return Pyproject{
		actionID:          actionID,
		spec:              s,
		rootDir:           dir,
		scmID:             scmID,
		versionFilter:     newFilter,
		uvAvailable:       isUvAvailable(),
		releaseAge:        releaseAge,
		trustedRegistries: trustedRegistries(s.PyPIMirrors),
	}, nil
}

// trustedRegistries returns the normalized URLs of pypi.org and the given mirrors.
func trustedRegistries(mirrors []string) map[string]bool {
	trusted := map[string]bool{normalizeRegistryURL(pypiRegistry): true}
	for _, mirror := range mirrors {
		trusted[normalizeRegistryURL(mirror)] = true
	}
	return trusted
}

// DiscoverManifests returns updatecli manifests for all Python dependencies found under rootDir.
func (p Pyproject) DiscoverManifests() ([][]byte, error) {
	log.Heading("Pyproject")

	return p.discoverDependencyManifests()
}
