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

// versionNumberRegex extracts the first PEP 440 version number from a constraint string.
// Used by isMatchingRules to evaluate Only/Ignore package version specifiers.
var versionNumberRegex = regexp.MustCompile(`\d+(\.\d+)*((?:a|b|rc)\d+)?`)

// depNameRegex parses a PEP 508 dependency string into name, optional extras, and constraint.
// Group 1: package name
// Group 4: extras (contents inside brackets, if any)
// Group 5: version constraint (remainder after name/extras)
var depNameRegex = regexp.MustCompile(`^([A-Za-z0-9]([A-Za-z0-9._-]*[A-Za-z0-9])?)(\[([^\]]+)\])?\s*(.*)$`)

// directRefRegex detects PEP 508 direct references (name @ <url>).
// Whitespace around @ is optional per the spec, so we match any amount of it.
var directRefRegex = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9._-]*[A-Za-z0-9])?\s*@\s*\S`)

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

	// Skip URL dependencies (PEP 508 direct references: name @ <url>).
	// Whitespace around @ is optional per the spec, so we use a regex rather than
	// a simple strings.Contains check.
	if directRefRegex.MatchString(dep) {
		return pythonDependency{}, fmt.Errorf("URL dependency not supported: %q", dep)
	}

	matches := depNameRegex.FindStringSubmatch(dep)
	if matches == nil {
		return pythonDependency{}, fmt.Errorf("could not parse dependency: %q", dep)
	}

	name := matches[1]
	extras := matches[4]
	constraint := strings.TrimSpace(matches[5])

	version := versionNumberRegex.FindString(constraint)

	return pythonDependency{
		Name:       name,
		Extras:     extras,
		Constraint: constraint,
		Version:    version,
	}, nil
}
