package precommit

import (
	"io/fs"
	"path/filepath"

	"github.com/updatecli/updatecli/pkg/core/log"
)

const (
	PrecommitConfigFile string = ".pre-commit-config.yaml"
	gittagIdentifier    string = "gittag"
)

// searchPrecommitConfigFiles looks, recursively, for every files named .pre-commit-config.yaml from a root directory.
func searchPrecommitConfigFiles(rootDir string) ([]string, error) {
	foundFiles := []string{}

	log.Debugf("Looking for Precommit configuration modules in %q", rootDir)

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Warningf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
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
