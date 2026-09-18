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
