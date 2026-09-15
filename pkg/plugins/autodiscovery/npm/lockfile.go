package npm

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"go.yaml.in/yaml/v3"
)

// lockedVersions holds the package versions resolved in a lock file.
type lockedVersions struct {
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

// loadLockedVersions returns the versions resolved by the lock file of a package.json directory.
// Only the lock files whose package manager is available are read, as the other ones prevent any update.
func loadLockedVersions(dir string, support lockFileSupport) lockedVersions {
	var lockFile string
	var parse func([]byte) (lockedVersions, error)

	switch {
	case support.npm:
		lockFile, parse = "package-lock.json", parsePackageLock
	case support.pnpm:
		lockFile, parse = "pnpm-lock.yaml", parsePnpmLock
	case support.yarn:
		lockFile, parse = "yarn.lock", parseYarnLock
	default:
		return lockedVersions{}
	}

	lockFile = filepath.Join(dir, lockFile)

	data, err := os.ReadFile(lockFile)
	if err != nil {
		logrus.Debugf("reading lock file %q: %s", lockFile, err)
		return lockedVersions{}
	}

	versions, err := parse(data)
	if err != nil {
		logrus.Debugf("parsing lock file %q: %s", lockFile, err)
		return lockedVersions{}
	}

	return versions
}

// parsePackageLock returns the top-level package versions of a package-lock.json file.
func parsePackageLock(data []byte) (lockedVersions, error) {
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

	versions := map[string]string{}

	for name, dependency := range lock.Dependencies {
		if dependency.Version != "" {
			versions[name] = dependency.Version
		}
	}

	for packagePath, dependency := range lock.Packages {
		name, found := strings.CutPrefix(packagePath, "node_modules/")
		// Nested packages are dependencies of dependencies
		if !found || strings.Contains(name, "/node_modules/") || dependency.Version == "" {
			continue
		}
		versions[name] = dependency.Version
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

// parsePnpmLock returns the direct dependency versions of the project next to a pnpm-lock.yaml file.
func parsePnpmLock(data []byte) (lockedVersions, error) {
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

	versions := map[string]string{}

	for _, project := range []pnpmProject{lock.pnpmProject, lock.Importers["."]} {
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
// It supports both Yarn classic (v1) and Yarn Berry (v2+) lock files, whose entries look like:
//
//	"axios@^1.0.0", axios@^1.1.0:     |  "axios@npm:^1.0.0, axios@npm:^1.1.0":
//	  version "1.2.6"                 |    version: 1.2.6
func parseYarnLock(data []byte) (lockedVersions, error) {
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
