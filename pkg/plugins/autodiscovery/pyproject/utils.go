package pyproject

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

const pyprojectFile = "pyproject.toml"

// skipDirs lists directories that should never be walked for pyproject.toml files.
var skipDirs = map[string]bool{
	".venv":        true,
	"__pycache__":  true,
	".git":         true,
	"node_modules": true,
	".tox":         true,
	".nox":         true,
	".eggs":        true,
}

// versionNumberRegex extracts the first PEP 440 version (including pre-release suffix) from a constraint.
// Matches: 0.51b0, 1.0a1, 2.0rc1, 2.31.0, 0.51
var versionNumberRegex = regexp.MustCompile(`\d+(\.\d+)*((?:a|b|rc)\d+)?`)

// pep440PreRelease matches PEP 440 pre-release tags for normalization to semver.
var pep440PreRelease = regexp.MustCompile(`^(\d+(?:\.\d+)*)(a|b|rc)(\d+)$`)

// normalizePEP440Version converts a PEP 440 version to semver-compatible form.
//
//	0.51b0  → 0.51.0-beta.0
//	1.0a1   → 1.0.0-alpha.1
//	2.0rc1  → 2.0.0-rc.1
//	2.31.0  → 2.31.0
//	0.51    → 0.51.0
func normalizePEP440Version(v string) string {
	m := pep440PreRelease.FindStringSubmatch(v)
	if m == nil {
		// Ensure at least 3-part version for semver
		if strings.Count(v, ".") < 2 && regexp.MustCompile(`^\d+(\.\d+)*$`).MatchString(v) {
			return v + ".0"
		}
		return v
	}

	base := m[1]
	if strings.Count(base, ".") < 2 {
		base += ".0"
	}

	switch m[2] {
	case "a":
		return base + "-alpha." + m[3]
	case "b":
		return base + "-beta." + m[3]
	case "rc":
		return base + "-rc." + m[3]
	}
	return v
}

// depNameRegex parses a PEP 508 dependency string into name, optional extras, and constraint.
// Group 1: package name
// Group 4: extras (contents inside brackets, if any)
// Group 5: version constraint (remainder after name/extras)
var depNameRegex = regexp.MustCompile(`^([A-Za-z0-9]([A-Za-z0-9._-]*[A-Za-z0-9])?)(\[([^\]]+)\])?\s*(.*)$`)

// findPyprojectFiles walks rootDir recursively and returns absolute paths to every
// pyproject.toml found, skipping common non-source directories.
func findPyprojectFiles(rootDir string) ([]string, error) {
	var found []string

	err := filepath.WalkDir(rootDir, func(path string, di fs.DirEntry, err error) error {
		if err != nil {
			logrus.Errorf("accessing path %q: %v", path, err)
			return err
		}

		if di.IsDir() {
			if skipDirs[di.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		if di.Name() == pyprojectFile {
			found = append(found, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	logrus.Debugf("%d pyproject.toml file(s) found", len(found))
	for _, f := range found {
		logrus.Debugf("    * %q", f)
	}

	return found, nil
}

// isUvAvailable reports whether the uv CLI is present on PATH.
func isUvAvailable() bool {
	return exec.Command("uv", "--version").Run() == nil
}

// isLockFileDetected reports whether the given lockfile path exists on disk.
func isLockFileDetected(lockfile string) bool {
	_, err := os.Stat(lockfile)
	return err == nil
}

// lockFileSupport describes which lock-file package managers are available.
type lockFileSupport struct {
	uv bool
	// poetry bool // future
	// pdm    bool // future
}

// detectLockFileSupport checks which lock-file managers are present for the given directory.
// Returns the support config and whether this pyproject.toml should be skipped entirely.
// Skipping happens when a lock file exists but the required CLI is not available —
// proceeding without the CLI would leave the lock file out of sync.
//
// uvAvailable is passed in (rather than calling isUvAvailable()) so that tests can
// control the value deterministically without requiring the real uv CLI on PATH.
func detectLockFileSupport(dir string, uvAvailable bool) (lockFileSupport, bool) {
	support := lockFileSupport{}

	if isLockFileDetected(filepath.Join(dir, "uv.lock")) {
		if uvAvailable {
			support.uv = true
		} else {
			logrus.Warning("skipping, uv.lock detected but Updatecli couldn't detect the uv command to update it in case of a pyproject.toml update")
			return support, true
		}
	}

	return support, false
}

// pythonDependency holds a parsed PEP 508 dependency string.
type pythonDependency struct {
	Name       string
	Extras     string // e.g. "jupyter"
	Constraint string // e.g. ">=2.28"
	Version    string // first version number extracted, e.g. "2.28"
}

// parsePEP508 parses a single PEP 508 dependency specifier.
// Returns an error for URL dependencies (@ https://...) which we cannot handle.
func parsePEP508(dep string) (pythonDependency, error) {
	// Strip environment markers: everything after " ; " or ";"
	if idx := strings.Index(dep, " ; "); idx != -1 {
		dep = dep[:idx]
	} else if idx := strings.Index(dep, ";"); idx != -1 {
		dep = dep[:idx]
	}
	dep = strings.TrimSpace(dep)

	// Skip URL dependencies (PEP 440 direct references: name @ https://...)
	if strings.Contains(dep, " @ ") {
		return pythonDependency{}, fmt.Errorf("URL dependency not supported: %q", dep)
	}

	matches := depNameRegex.FindStringSubmatch(dep)
	if matches == nil {
		return pythonDependency{}, fmt.Errorf("could not parse dependency: %q", dep)
	}

	name := matches[1]
	extras := matches[4]
	constraint := strings.TrimSpace(matches[5])

	// Extract the first PEP 440 version from the constraint and normalize to semver.
	rawVersion := versionNumberRegex.FindString(constraint)
	version := normalizePEP440Version(rawVersion)

	return pythonDependency{
		Name:       name,
		Extras:     extras,
		Constraint: constraint,
		Version:    version,
	}, nil
}
