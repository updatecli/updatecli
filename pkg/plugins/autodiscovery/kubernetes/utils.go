package kubernetes

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	goyaml "gopkg.in/yaml.v3"
)

// searchKubernetesFiles will look, recursively, for every files with an extension .yaml or .yml from a root directory.
func searchKubernetesFiles(rootDir string, files []string) ([]string, error) {
	kubernetesFiles := []string{}

	logrus.Debugf("Looking for Kubernetes file(s) in %q", rootDir)

	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		for _, f := range files {
			if !info.IsDir() {
				match, err := filepath.Match(f, info.Name())
				if err != nil {
					continue
				}

				if match {
					kubernetesFiles = append(kubernetesFiles, path)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	logrus.Debugf("%d Kubernetes file(s) found", len(kubernetesFiles))
	for _, foundFile := range kubernetesFiles {
		kubernetesFile := filepath.Base(foundFile)
		logrus.Debugf("    * %q", kubernetesFile)
	}

	return kubernetesFiles, nil
}

// getManifestData reads a T file for information that could be automatically updated.
func getManifestData[T any](filename, logPrefix string) (*T, error) {
	data := new(T)

	manifestFile := filepath.Base(filepath.Dir(filename))
	logrus.Debugf("%s file %q found in %q", logPrefix, manifestFile, filepath.Dir(filename))

	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	err = goyaml.Unmarshal(content, data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// getProwManifestData reads a Prow file for information that could be automatically updated.
func getProwManifestData(filename string) (*prowFlavorManifestSpec, error) {
	return getManifestData[prowFlavorManifestSpec](filename, "Prow")
}

// getKubernetesManifestData reads a Kubernetes file for information that could be automatically updated.
func getKubernetesManifestData(filename string) (*kubernetesFlavorManifestSpec, error) {
	return getManifestData[kubernetesFlavorManifestSpec](filename, "Kubernetes")
}
