package precommit

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

const (
	PrecommitConfigFile string = ".pre-commit-config.yaml"
)

// searchPrecommitConfigFiles looks, recursively, for every files named .pre-commit-config.yaml from a root directory.
func searchPrecommitConfigFiles(rootDir string) ([]string, error) {
	foundFiles := []string{}

	logrus.Debugf("Looking for Precommit configuration modules in %q", rootDir)

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		fileName := d.Name()

		if fileName == PrecommitConfigFile {
			foundFiles = append(foundFiles, path)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return foundFiles, nil
}
