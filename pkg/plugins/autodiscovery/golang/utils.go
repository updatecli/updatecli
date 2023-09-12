package golang

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

const (
	GoModFile string = "go.mod"
)

// searchGoModFiles looks, recursively, for every files named go.mod from a root directory.
func searchGoModFiles(rootDir string) ([]string, error) {

	foundFiles := []string{}

	// To do switch to WalkDir which is more efficient, introduced in 1.16
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if info.Name() == GoModFile {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return fmt.Errorf("get absolute path of %q: %s", path, err)
			}
			foundFiles = append(foundFiles, absPath)
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

func getGoModContent(filename string) (goVersion string, goModules map[string]string, err error) {

	data, err := os.ReadFile(filename)

	if err != nil {
		return goVersion, goModules, err
	}

	modfile, err := modfile.Parse(filename, data, nil)
	if err != nil {
		return goVersion, goModules, err
	}

	goVersion = modfile.Go.Version

	goModules = make(map[string]string)
	for _, r := range modfile.Require {
		if !r.Indirect {
			goModules[r.Mod.Path] = r.Mod.Version
		}
	}

	return goVersion, goModules, nil
}
