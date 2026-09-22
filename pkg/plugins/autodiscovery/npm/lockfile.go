package npm

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"go.yaml.in/yaml/v3"
)

// lockedVersions holds the package versions resolved in a lock file.
type lockedVersions struct {
	// lockFile holds the path of the lock file the versions come from, empty when no lock file was found.
	lockFile string
	// byName maps package names to their resolved version, as recorded by npm and pnpm.
	byName map[string]string
	// byDescriptor maps dependency descriptors, such as "axios@^1.0.0", to their resolved version, as recorded by Yarn.
	byDescriptor map[string]string
}

// version returns the version resolved for a package.json dependency, or an empty string when unknown.
func (l lockedVersions) version(name, constraint string) string {
	if version, ok := l.byDescriptor[name+"@"+constraint]; ok {
		return version
	}
	return l.byName[name]
}

// errUnknownImporter reports a lock file that doesn't record the project it was looked up for,
// such as the lock file of an unrelated project in a parent directory.
var errUnknownImporter = errors.New("project not recorded in the lock file")

// lockFileParser returns the versions a lock file resolved for one of the projects it records,
// identified by its path relative to the lock file, such as "." or "packages/app".
// It returns errUnknownImporter when the lock file doesn't record that project.
type lockFileParser func(lockFile string, data []byte, importer string) (lockedVersions, error)

// lockFiles associates the supported lock files with their parser, in the order they are looked up.
var lockFiles = []struct {
	name  string
	parse lockFileParser
}{
	{name: "package-lock.json", parse: parsePackageLock},
	{name: "pnpm-lock.yaml", parse: parsePnpmLock},
	{name: "yarn.lock", parse: parseYarnLock},
}

// loadLockedVersions returns the versions resolved by the lock file of a package.json directory.
// A missing lock file returns no version and no error, while an unreadable or malformed one returns an
// error, as silently ignoring it would drop every constrained dependency from the vulnerability report.
func loadLockedVersions(dir, rootDir string) (lockedVersions, error) {
	lockFile, parse := searchLockFile(dir, rootDir)
	if parse == nil {
		logrus.Debugf("no lock file found for %q", dir)
		return lockedVersions{}, nil
	}

	importer, err := filepath.Rel(filepath.Dir(lockFile), dir)
	if err != nil {
		return lockedVersions{}, fmt.Errorf("locating %q from lock file %q: %w", dir, lockFile, err)
	}

	data, err := os.ReadFile(lockFile)
	if err != nil {
		return lockedVersions{}, fmt.Errorf("reading lock file %q: %w", lockFile, err)
	}

	versions, err := parse(lockFile, data, filepath.ToSlash(importer))
	if errors.Is(err, errUnknownImporter) {
		// The lock file belongs to another project, so it says nothing about the versions installed for this one
		logrus.Debugf("%q is not a project of lock file %q", dir, lockFile)
		return lockedVersions{}, nil
	}
	if err != nil {
		return lockedVersions{}, fmt.Errorf("parsing lock file %q: %w", lockFile, err)
	}
	versions.lockFile = lockFile

	return versions, nil
}

// searchLockFile returns the closest lock file of a package.json directory, looking up to the root
// directory the search started from, as a workspace only holds a lock file at its root.
// A lock file found in a parent directory only applies if it records the project, which its parser checks.
// The package manager doesn't have to be available to read the versions it resolved, unlike to update them.
func searchLockFile(dir, rootDir string) (string, lockFileParser) {
	dir, rootDir = filepath.Clean(dir), filepath.Clean(rootDir)

	if relative, err := filepath.Rel(rootDir, dir); err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		// The package.json is outside of the searched directory, so only its own directory can be looked up
		rootDir = dir
	}

	for {
		for _, candidate := range lockFiles {
			lockFile := filepath.Join(dir, candidate.name)
			if isLockFileDetected(lockFile) {
				return lockFile, candidate.parse
			}
		}

		parent := filepath.Dir(dir)
		if dir == rootDir || parent == dir {
			return "", nil
		}
		dir = parent
	}
}

// parsePackageLock returns the package versions a package-lock.json file installed for a project,
// either hoisted at its root or, for a workspace project, next to it.
func parsePackageLock(_ string, data []byte, importer string) (lockedVersions, error) {
	type lockedPackage struct {
		Version string `json:"version"`
	}

	var lock struct {
		// Packages lists installed packages by path, since lockfileVersion 2.
		Packages map[string]lockedPackage `json:"packages"`
		// Dependencies lists top-level packages by name, up to lockfileVersion 2.
		Dependencies map[string]lockedPackage `json:"dependencies"`
	}

	if err := json.Unmarshal(data, &lock); err != nil {
		return lockedVersions{}, err
	}

	// Workspace projects are recorded by path, since lockfileVersion 2
	if _, found := lock.Packages[importer]; importer != "." && !found {
		return lockedVersions{}, errUnknownImporter
	}

	versions := map[string]string{}

	for name, dependency := range lock.Dependencies {
		if dependency.Version != "" {
			versions[name] = dependency.Version
		}
	}

	prefixes := []string{"node_modules/"}
	if importer != "." {
		// The packages installed next to a workspace project take precedence over the hoisted ones
		prefixes = append(prefixes, importer+"/node_modules/")
	}

	for _, prefix := range prefixes {
		for packagePath, dependency := range lock.Packages {
			name, found := strings.CutPrefix(packagePath, prefix)
			// Nested packages are dependencies of dependencies
			if !found || strings.Contains(name, "/node_modules/") || dependency.Version == "" {
				continue
			}
			versions[name] = dependency.Version
		}
	}

	return lockedVersions{byName: versions}, nil
}

// pnpmDependency is a dependency of a pnpm-lock.yaml file, recorded either as a version,
// up to lockfileVersion 5, or as a specifier and a version, since lockfileVersion 6.
type pnpmDependency struct {
	Version string
}

func (d *pnpmDependency) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.ScalarNode {
		d.Version = value.Value
		return nil
	}

	var dependency struct {
		Version string `yaml:"version"`
	}
	if err := value.Decode(&dependency); err != nil {
		return err
	}
	d.Version = dependency.Version

	return nil
}

// parsePnpmLock returns the direct dependency versions of one project of a pnpm-lock.yaml file.
func parsePnpmLock(_ string, data []byte, importer string) (lockedVersions, error) {
	type pnpmProject struct {
		Dependencies    map[string]pnpmDependency `yaml:"dependencies"`
		DevDependencies map[string]pnpmDependency `yaml:"devDependencies"`
	}

	var lock struct {
		// Importers lists the workspace projects by path, since lockfileVersion 6 for every project.
		Importers map[string]pnpmProject `yaml:"importers"`
		// The inline project is used by single projects up to lockfileVersion 6.
		pnpmProject `yaml:",inline"`
	}

	if err := yaml.Unmarshal(data, &lock); err != nil {
		return lockedVersions{}, err
	}

	project, found := lock.Importers[importer]
	if importer != "." && !found {
		return lockedVersions{}, errUnknownImporter
	}

	projects := []pnpmProject{project}
	if importer == "." {
		projects = []pnpmProject{lock.pnpmProject, lock.Importers["."]}
	}

	versions := map[string]string{}

	for _, project := range projects {
		for _, dependencies := range []map[string]pnpmDependency{project.Dependencies, project.DevDependencies} {
			for name, dependency := range dependencies {
				// Remove the peer dependencies suffix, such as "1.0.0(react@18.2.0)" or "1.0.0_react@18.2.0"
				version, _, _ := strings.Cut(dependency.Version, "(")
				version, _, _ = strings.Cut(version, "_")
				if version != "" {
					versions[name] = version
				}
			}
		}
	}

	return lockedVersions{byName: versions}, nil
}

// parseYarnLock returns the resolved version of each dependency descriptor of a yarn.lock file.
// Yarn records the descriptors of every workspace project in a single lock file, so they are all returned
// once the project is known to be a workspace of the lock file.
// It supports both Yarn classic (v1) and Yarn Berry (v2+) lock files, whose entries look like:
//
//	"axios@^1.0.0", axios@^1.1.0:     |  "axios@npm:^1.0.0, axios@npm:^1.1.0":
//	  version "1.2.6"                 |    version: 1.2.6
func parseYarnLock(lockFile string, data []byte, importer string) (lockedVersions, error) {
	if importer != "." && !isYarnWorkspace(filepath.Dir(lockFile), importer) {
		return lockedVersions{}, errUnknownImporter
	}

	versions := map[string]string{}
	var descriptors []string

	for line := range strings.Lines(string(data)) {
		line = strings.TrimRight(line, "\r\n ")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Entry header
		if !strings.HasPrefix(line, " ") {
			descriptors = descriptors[:0]
			for descriptor := range strings.SplitSeq(strings.TrimSuffix(line, ":"), ",") {
				descriptors = append(descriptors, strings.Trim(strings.TrimSpace(descriptor), `"`))
			}
			continue
		}

		// Entry fields are indented by two spaces, deeper lines belong to nested fields such as dependencies
		if strings.HasPrefix(line, "   ") {
			continue
		}

		key, value, found := strings.Cut(strings.TrimSpace(line), " ")
		if !found || strings.TrimSuffix(key, ":") != "version" {
			continue
		}

		version := strings.Trim(strings.TrimSpace(value), `"`)
		for _, descriptor := range descriptors {
			versions[normalizeYarnDescriptor(descriptor)] = version
		}
	}

	return lockedVersions{byDescriptor: versions}, nil
}

// isYarnWorkspace reports whether a project, identified by its path relative to a directory, matches one of
// the workspaces declared by the package.json of that directory, as yarn.lock doesn't list workspaces.
func isYarnWorkspace(dir, importer string) bool {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		logrus.Debugf("reading workspaces of %q: %s", dir, err)
		return false
	}

	var manifest struct {
		Workspaces json.RawMessage `json:"workspaces"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil || len(manifest.Workspaces) == 0 {
		return false
	}

	// Workspaces are declared as a list of patterns, or in Yarn classic as an object holding them
	var patterns []string
	if err := json.Unmarshal(manifest.Workspaces, &patterns); err != nil {
		var workspaces struct {
			Packages []string `json:"packages"`
		}
		if err := json.Unmarshal(manifest.Workspaces, &workspaces); err != nil {
			return false
		}
		patterns = workspaces.Packages
	}

	for _, pattern := range patterns {
		if matched, err := path.Match(path.Clean(pattern), importer); err == nil && matched {
			return true
		}
	}

	return false
}

// normalizeYarnDescriptor removes the Yarn Berry "npm:" protocol, so "axios@npm:^1.0.0" matches
// the package.json dependency "axios" with the constraint "^1.0.0".
func normalizeYarnDescriptor(descriptor string) string {
	// Scoped package names, such as "@mdi/font", start with "@"
	separator := strings.Index(descriptor[min(1, len(descriptor)):], "@") + 1
	if separator == 0 {
		return descriptor
	}

	return descriptor[:separator] + "@" + strings.TrimPrefix(descriptor[separator+1:], "npm:")
}
