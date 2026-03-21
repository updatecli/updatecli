package flux

import (
	"io/fs"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// searchFluxFiles will look, recursively, for every files containing a flux helmrelease or helmrepository from a root directory.
func (f *Flux) searchFluxFiles(rootDir string, files []string) error {

	logrus.Debugf("Looking for Flux file(s) in %q", rootDir)

	if err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			logrus.Debugf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
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

				ociRepository, err := loadOCIRepository(path)
				if err != nil {
					logrus.Debugf("Failed loading document from %s as OCI Repository: %s", path, err)
				}
				if len(ociRepository) > 0 {
					f.ociRepositoryFiles = append(f.ociRepositoryFiles, path)
				}

				helmRepository, err := loadHelmRepositoryData(path)
				if err != nil {
					logrus.Debugf("Failed loading document from %s as HelmRepository: %s", path, err)
				}
				if len(helmRepository) > 0 {
					f.helmRepositories = append(f.helmRepositories, helmRepository...)
				}

				helmRelease, err := loadHelmRelease(path)
				if err != nil {
					logrus.Debugf("Failed loading document from %s as HelmRelease: %s", path, err)
				}
				if len(helmRelease) > 0 {
					logrus.Debugf("Found %d HelmRelease object(s) in %s", len(helmRelease), path)
					f.helmReleaseFiles = append(f.helmReleaseFiles, path)
				}
			}
		}

		return nil
	}); err != nil {
		return err
	}

	if len(f.helmReleaseFiles) > 0 {
		logrus.Debugf("%d Flux file(s) found container helmRelease definition", len(f.helmReleaseFiles))
	}
	if len(f.ociRepositoryFiles) > 0 {
		logrus.Debugf("%d Flux file(s) found container ociRepository definition", len(f.ociRepositoryFiles))
	}
	if len(f.helmRepositories) > 0 {
		logrus.Debugf("%d Flux file(s) found container helmRepository definition", len(f.helmRepositories))
	}

	return nil
}
