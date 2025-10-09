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
)

const (
	GoModFile string = "go.mod"
)

type Replace struct {
	OldPath    string
	OldVersion string
	NewPath    string
	NewVersion string
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

func getGoModContent(filename string) (goVersion string, goModules map[string]string, replaceGoModules []Replace, err error) {

	data, err := os.ReadFile(filename)

	if err != nil {
		return "", nil, nil, err
	}

	modfile, err := modfile.Parse(filename, data, nil)
	if err != nil {
		return "", nil, nil, err
	}

	goVersion = modfile.Go.Version

	for _, r := range modfile.Require {
		if !r.Indirect {
			if goModules == nil {
				goModules = make(map[string]string)
			}
			goModules[r.Mod.Path] = r.Mod.Version
		}
	}

	for _, r := range modfile.Replace {
		// Ignore replace directives with local path
		if strings.HasPrefix(r.New.Path, ".") || strings.HasPrefix(r.New.Path, "/") {
			continue
		}
		replaceGoModules = append(replaceGoModules, Replace{
			OldPath:    r.Old.Path,
			OldVersion: r.Old.Version,
			NewPath:    r.New.Path,
			NewVersion: r.New.Version,
		})
	}

	return goVersion, goModules, replaceGoModules, nil
}
