package pyproject

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/BurntSushi/toml"
)

const uvLockFile = "uv.lock"

// pypiRegistry is the registry uv.lock records for packages resolved from pypi.org.
const pypiRegistry = "https://pypi.org/simple"

// exactPinRegex matches a constraint pinning one exact version, such as "==1.2.3".
// Wildcards ("==1.*") and arbitrary equality ("===") are not exact pins.
var exactPinRegex = regexp.MustCompile(`^==\s*([0-9][0-9A-Za-z.+!-]*)$`)

// exactPinnedVersion returns the version a constraint pins exactly, or "" when it allows a range.
func exactPinnedVersion(constraint string) string {
	m := exactPinRegex.FindStringSubmatch(strings.TrimSpace(constraint))
	if m == nil {
		return ""
	}
	return m[1]
}

// uvLock mirrors the subset of uv.lock we care about.
type uvLock struct {
	Package []struct {
		Name    string `toml:"name"`
		Version string `toml:"version"`
		// Source tells where uv resolved the package from, such as
		// { registry = "https://pypi.org/simple" }, { workspace = true } or { git = "..." }.
		Source map[string]interface{} `toml:"source"`
	} `toml:"package"`
}

// lockedVersions maps a normalized package name to the versions uv.lock resolved for it.
// A package resolved at more than one version comes from uv forking on environment markers.
type lockedVersions map[string][]string

// lockedPackages holds what uv.lock resolved, split by where it resolved packages from.
type lockedPackages struct {
	// versions holds the packages resolved from pypi.org or one of its trusted mirrors.
	versions lockedVersions
	// notFromRegistry holds the normalized names resolved from a workspace, a path,
	// a git repository or a URL, which are not PyPI releases.
	notFromRegistry map[string]bool
	// untrustedRegistry maps the normalized names resolved from another registry, such as a
	// private index, to that registry. The OSV database only covers pypi.org, so a private
	// package sharing the name of a public one would get its advisories.
	untrustedRegistry map[string]string
}

// normalizeRegistryURL returns a registry URL in the form used to compare registries.
func normalizeRegistryURL(registry string) string {
	return strings.TrimRight(strings.TrimSpace(registry), "/")
}

// loadUvLockVersions reads the uv.lock in dir. trustedRegistries holds the normalized URLs
// of the registries serving PyPI releases: pypi.org and the mirrors the user trusts.
// Only packages resolved from those are kept in versions. The others are returned apart,
// so that they aren't checked through a version pinned in pyproject.toml either.
func loadUvLockVersions(dir string, trustedRegistries map[string]bool) (lockedPackages, error) {
	var data uvLock

	lockFile := filepath.Join(dir, uvLockFile)
	if _, err := toml.DecodeFile(lockFile, &data); err != nil {
		return lockedPackages{}, fmt.Errorf("parsing %q: %w", lockFile, err)
	}

	locked := lockedPackages{
		versions:          lockedVersions{},
		notFromRegistry:   map[string]bool{},
		untrustedRegistry: map[string]string{},
	}
	for _, pkg := range data.Package {
		if pkg.Name == "" {
			continue
		}
		name := normalizeName(pkg.Name)

		rawRegistry, fromRegistry := pkg.Source["registry"]
		if !fromRegistry {
			locked.notFromRegistry[name] = true
			continue
		}
		registry, _ := rawRegistry.(string)
		if !trustedRegistries[normalizeRegistryURL(registry)] {
			locked.untrustedRegistry[name] = registry
			continue
		}

		if pkg.Version == "" {
			continue
		}
		locked.versions[name] = append(locked.versions[name], pkg.Version)
	}

	return locked, nil
}
