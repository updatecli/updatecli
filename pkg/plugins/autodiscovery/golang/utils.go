package golang

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"golang.org/x/mod/semver"
)

const (
	GoModFile string = "go.mod"
)

type Replace struct {
	OldPath    string
	OldVersion string
	NewPath    string
	NewVersion string
	// Ambiguous indicates that the replace directive has no version while another replace directive
	// with a version exists for the same module, so the "golang/gomod" target can't select it
	Ambiguous bool
}

// isLocal returns true if the replacement is a local path
func (r Replace) isLocal() bool {
	return isLocalPath(r.NewPath)
}

func isLocalPath(path string) bool {
	return strings.HasPrefix(path, ".") || strings.HasPrefix(path, "/")
}

// searchGoModFiles looks, recursively, for every files named go.mod from a root directory.
func searchGoModFiles(rootDir string) ([]string, error) {

	foundFiles := []string{}

	logrus.Debugf("Looking for Go mod file(s) in %q", rootDir)

	// To do switch to WalkDir which is more efficient, introduced in 1.16
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if info.Name() == GoModFile {
			foundFiles = append(foundFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return foundFiles, nil
}

func isGolangInstalled() bool {
	cmd := exec.Command("go", "version")
	err := cmd.Run()
	return err == nil
}

// getGoModContent parses a go.mod file and returns its go version, its direct modules,
// the replace directives pointing to a remote module, and for each replaced direct module
// the replace directive applied by Go, including the ones pointing to a local path.
func getGoModContent(filename string) (goVersion string, goModules map[string]string, replaceGoModules []Replace, appliedReplaces map[string]Replace, err error) {

	data, err := os.ReadFile(filename)

	if err != nil {
		return "", nil, nil, nil, err
	}

	f, err := modfile.Parse(filename, data, nil)
	if err != nil {
		return "", nil, nil, nil, err
	}

	goVersion = f.Go.Version

	for _, r := range f.Require {
		if !r.Indirect {
			if goModules == nil {
				goModules = make(map[string]string)
			}
			goModules[r.Mod.Path] = r.Mod.Version
		}
	}

	// A replace directive matching the required version takes precedence over
	// a replace directive without version, which applies to every version of the module.
	applied := make(map[string]*modfile.Replace)
	versioned := make(map[string]bool)
	for _, r := range f.Replace {
		if r.Old.Version != "" {
			versioned[r.Old.Path] = true
		}

		version, found := goModules[r.Old.Path]
		if !found {
			continue
		}

		switch r.Old.Version {
		case version:
			applied[r.Old.Path] = r
		case "":
			if _, found := applied[r.Old.Path]; !found {
				applied[r.Old.Path] = r
			}
		}
	}

	for path, r := range applied {
		if appliedReplaces == nil {
			appliedReplaces = make(map[string]Replace)
		}
		appliedReplaces[path] = Replace{
			OldPath:    r.Old.Path,
			OldVersion: r.Old.Version,
			NewPath:    r.New.Path,
			NewVersion: r.New.Version,
			Ambiguous:  r.Old.Version == "" && versioned[path],
		}
	}

	for _, r := range f.Replace {
		// Ignore replace directives with local path
		if isLocalPath(r.New.Path) {
			continue
		}
		replaceGoModules = append(replaceGoModules, Replace{
			OldPath:    r.Old.Path,
			OldVersion: r.Old.Version,
			NewPath:    r.New.Path,
			NewVersion: r.New.Version,
		})
	}

	return goVersion, goModules, replaceGoModules, appliedReplaces, nil
}

// isPseudoVersion checks if the provided version is a pseudo-version.
func isPseudoVersion(version string) bool {
	return module.IsPseudoVersion(version) || module.IsZeroPseudoVersion(version)
}

// moduleVersionPattern returns the version filter kind and pattern used to look for a version newer than the provided one.
func (g Golang) moduleVersionPattern(version string) (kind, pattern string, err error) {
	kind = g.versionFilter.Kind

	pattern, err = g.versionFilter.GreaterThanPattern(version)
	if err != nil {
		return "", "", err
	}

	if !isPseudoVersion(version) || kind != "semver" {
		return kind, pattern, nil
	}

	/*
		A pseudo version is a semver prerelease, so patterns such as "patch" ("1.2.x-0")
		or a custom constraint such as "~1.2" accept tags older than the pseudo version.
		Each group of the constraint gets the pseudo version as lower bound, unless it
		already starts from it, so that no tag can downgrade the module.
	*/
	lowerBound := strings.TrimPrefix(version, "v")

	/*
		For a prerelease, GreaterThanPattern turns "majoronly" into "any version newer
		than the current one", which would also accept minor and patch updates.
		Requiring a greater major version keeps the update major only.
	*/
	if g.versionFilter.Pattern == "majoronly" {
		pattern = ">" + strings.TrimPrefix(semver.Major(version), "v")
	}

	groups := strings.Split(pattern, "||")
	for i, group := range groups {
		group = strings.TrimSpace(group)
		if !strings.Contains(group, lowerBound) {
			group = fmt.Sprintf(">=%s, %s", lowerBound, group)
		}
		groups[i] = group
	}
	pattern = strings.Join(groups, " || ")

	return kind, pattern, nil
}
