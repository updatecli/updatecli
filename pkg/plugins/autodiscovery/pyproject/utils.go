package pyproject

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	pep440 "github.com/aquasecurity/go-pep440-version"
	"github.com/updatecli/updatecli/pkg/core/log"
)

const (
	// pyprojectFile is the name of the files the autodiscovery looks for.
	pyprojectFile = "pyproject.toml"
	// defaultMinimumReleaseAge defines how long a release must have been published before
	// Updatecli suggests it. It reduces the exposure to malicious or broken releases that get
	// yanked or deleted shortly after. It is longer than npm's three days because PyPI has no
	// unpublish window: a release can be removed at any time, so detection time is what matters.
	defaultMinimumReleaseAge = "7d"
)

var (
	// relativeSemverPatterns lists the semver versionfilter patterns relative to a version.
	relativeSemverPatterns = map[string]bool{
		"prerelease": true,
		"patch":      true,
		"minor":      true,
		"minoronly":  true,
		"major":      true,
		"majoronly":  true,
	}

	// dependencyGroupNameRegex matches the dependency group names PEP 735 allows.
	dependencyGroupNameRegex = regexp.MustCompile(`(?i)^([A-Z0-9]|[A-Z0-9][A-Z0-9._-]*[A-Z0-9])$`)

	// skipDirs lists directories that should never be walked for pyproject.toml files.
	skipDirs = map[string]bool{
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
	versionNumberRegex = regexp.MustCompile(`\d+(\.\d+)*((?:a|b|rc)\d+)?`)

	// depNameRegex parses a PEP 508 dependency string into name, optional extras, and constraint.
	// Group 1: package name
	// Group 4: extras (contents inside brackets, if any)
	// Group 5: version constraint (remainder after name/extras)
	depNameRegex = regexp.MustCompile(`^([A-Za-z0-9]([A-Za-z0-9._-]*[A-Za-z0-9])?)(\[([^\]]+)\])?\s*(.*)$`)

	// directRefRegex detects PEP 508 direct references (name @ <url>).
	// Whitespace around @ is optional per the spec, so we match any amount of it.
	directRefRegex = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9._-]*[A-Za-z0-9])?\s*@\s*\S`)
)

// findPyprojectFiles walks rootDir recursively and returns absolute paths to every
// pyproject.toml found, skipping common non-source directories.
func findPyprojectFiles(rootDir string) ([]string, error) {
	var found []string

	err := filepath.WalkDir(rootDir, func(path string, di fs.DirEntry, err error) error {
		if err != nil {
			log.Errorf("accessing path %q: %v", path, err)
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

	log.Debugf("%d pyproject.toml file(s) found", len(found))
	for _, f := range found {
		log.Debugf("    * %q", f)
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
	// uv reports whether uv.lock is present and the uv command can update it.
	uv bool
	// uvLock reports whether uv.lock is present, even when the uv command is missing.
	uvLock bool
	// poetry bool // future
	// pdm    bool // future
}

// detectLockFileSupport checks which lock-file managers are present for the given directory.
// Returns the support config and whether this pyproject.toml should be skipped entirely.
// Skipping happens when a lock file exists but the required CLI is not available —
// proceeding without the CLI would leave the lock file out of sync. The caller reports it,
// since vulnerability mode can still read the lock file to report vulnerable packages.
//
// uvAvailable is passed in (rather than calling isUvAvailable()) so that tests can
// control the value deterministically without requiring the real uv CLI on PATH.
func detectLockFileSupport(dir string, uvAvailable bool) (lockFileSupport, bool) {
	support := lockFileSupport{}

	if isLockFileDetected(filepath.Join(dir, uvLockFile)) {
		support.uvLock = true
		if uvAvailable {
			support.uv = true
		} else {
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
	// PEP 508 also allows the constraint between parentheses, as in "requests (==2.25.0)".
	if inner, found := strings.CutPrefix(constraint, "("); found {
		if inner, found = strings.CutSuffix(inner, ")"); found {
			constraint = strings.TrimSpace(inner)
		}
	}

	version := versionNumberRegex.FindString(constraint)

	return pythonDependency{
		Name:       name,
		Extras:     extras,
		Constraint: constraint,
		Version:    version,
	}, nil
}

// declaredLowerBound returns the release a PEP 440 constraint starts from, such as "2.28" for
// ">=2.28,<3" or ">=2.28.post1", and "1" for "==1.*", or "" when it sets no lower bound, such as
// "<3.0" or "!=1.5". The release leaves out pre, post, dev and local segments, which a semantic
// version can't hold.
func declaredLowerBound(constraint string) string {
	constraint = strings.TrimSpace(constraint)
	for _, specifier := range strings.Split(constraint, ",") {
		specifier = strings.TrimSpace(specifier)
		// Longer operators first, so ">=" isn't taken for ">"
		for _, operator := range []string{"===", "==", "~=", ">=", ">"} {
			if rest, found := strings.CutPrefix(specifier, operator); found {
				bound := strings.TrimSuffix(strings.TrimSpace(rest), ".*")
				if v, err := pep440.Parse(bound); err == nil {
					return v.BaseVersion()
				}
				return bound
			}
		}
	}
	return ""
}
