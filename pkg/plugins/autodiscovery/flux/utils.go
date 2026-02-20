package flux

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// searchFluxFiles will look, recursively, for every files containing a flux helmrelease or helmrepository from a root directory.
func (f *Flux) searchFluxFiles(rootDir string, files []string) error {

	logrus.Debugf("Looking for Flux file(s) in %q", rootDir)

	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
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

				data, err := os.ReadFile(path)
				if err != nil {
					logrus.Debugf("Failed reading file %s: %s", path, err)
					continue
				}

				// Split YAML documents
				docs := bytes.Split(data, []byte("---"))

				// Track if we've found a HelmRelease in this file to avoid adding the path multiple times
				hasHelmRelease := false
				hasOCIRepository := false

				for _, doc := range docs {
					if len(bytes.TrimSpace(doc)) == 0 {
						continue
					}

					// Process each document separately
					helmRepository, err := loadHelmRepositoryFromBytes(doc)
					if err != nil {
						logrus.Debugf("Failed loading document from %s as HelmRepository: %s", path, err)
					}

					if helmRepository != nil {
						f.helmRepositories = append(f.helmRepositories, *helmRepository)
						continue
					}

					ociRepository, err := loadOCIRepositoryFromBytes(doc)
					if err != nil {
						logrus.Debugf("Failed loading document from %s as OCIRepository: %s", path, err)
					}

					if ociRepository != nil {
						if !hasOCIRepository {
							f.ociRepositoryFiles = append(f.ociRepositoryFiles, path)
							hasOCIRepository = true
						}
						continue
					}

					helmRelease, err := loadHelmReleaseFromBytes(doc)
					if err != nil {
						logrus.Debugf("Failed loading document from %s as HelmRelease: %s", path, err)
					}

					if helmRelease != nil && !hasHelmRelease {
						f.helmReleaseFiles = append(f.helmReleaseFiles, path)
						hasHelmRelease = true
					}
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
