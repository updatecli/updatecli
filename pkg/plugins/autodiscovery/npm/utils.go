package npm

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

const (
	PackageJsonFile string = "package.json"
)

// searchPackageJsonFiles looks, recursively, for every files named package.json from a root directory.
func searchPackageJsonFiles(rootDir string) ([]string, error) {

	foundFiles := []string{}

	// To do switch to WalkDir which is more efficient, introduced in 1.16
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		// Updatecli should ignore all package.json from the directory named "node_modules"
		// as they are automatically installed by npm
		if info.IsDir() && info.Name() == "node_modules" {
			return filepath.SkipDir
		}

		if info.Name() == PackageJsonFile {
			foundFiles = append(foundFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return foundFiles, nil
}
