package flux

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// searchFluxFiles will look, recursively, for every files containing a flux helmrelease or helmrepository from a root directory.
func (f *Flux) searchFluxFiles(rootDir string, files []string) error {

	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		for _, foundFile := range files {
			if !info.IsDir() {
				match, err := filepath.Match(foundFile, info.Name())
				if err != nil {
					continue
				}

				// if file doesn't match the pattern, skip it
				if !match {
					continue
				}

				// First try to see if our file contains a HelmRepository definition
				// if it does, then we save it and move to the next file path
				helmRepository, err := isHelmRepository(path)
				if err != nil {
					logrus.Debugf("Failed loading file %s as HelmRepository: %s", path, err.Error())
				}

				if helmRepository != nil {
					f.helmRepositories = append(f.helmRepositories, *helmRepository)
				}

				// Second, try to see if our file contains a OCIRepository definition
				// if it does, then we save it and move to the next file path
				ociRepository, err := loadOCIRepository(path)
				if err != nil {
					logrus.Debugf("Failed loading file %s as OCIRepository: %s", path, err.Error())
				}

				if ociRepository != nil {
					f.ociRepositoryFiles = append(f.ociRepositoryFiles, path)
				}

				// Third, try to see if our file contains a HelmRelease definition
				// if it does, then we save it and move to the next file path
				helmRelease, err := loadHelmRelease(path)
				if err != nil {
					logrus.Debugf("Failed loading file %s as HelmRelease: %s", path, err.Error())
				}

				if helmRelease != nil {
					f.helmReleaseFiles = append(f.helmReleaseFiles, path)
				}
			}
		}

		return nil
	})

	if err != nil {
		return err
	}

	logrus.Debugf("%d Flux file(s) found container helmRelease definition", len(f.helmReleaseFiles))
	logrus.Debugf("%d Flux file(s) found container ociRepository definition", len(f.ociRepositoryFiles))

	return nil
}
