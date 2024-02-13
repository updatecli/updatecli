package kubernetes

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	goyaml "gopkg.in/yaml.v3"
)

// searchKubernetesFiles will look, recursively, for every files with an extension .yaml or .yml from a root directory.
func searchKubernetesFiles(rootDir string, files []string) ([]string, error) {

	kubernetesFiles := []string{}

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
		kubernetesFile := filepath.Base(filepath.Dir(foundFile))
		logrus.Debugf("    * %q", kubernetesFile)
	}

	return kubernetesFiles, nil
}

// getKubernetesManifestData reads a Kubernetes file for information that could be automatically updated.
func getKubernetesManifestData(filename string) (*kubernetesManifestSpec, error) {

	var kubernetesData kubernetesManifestSpec

	kubernetesFile := filepath.Base(filepath.Dir(filename))
	logrus.Debugf("Kubernetes file %q found in %q", kubernetesFile, filepath.Dir(filename))

	if _, err := os.Stat(filename); err != nil {
		return &kubernetesManifestSpec{}, err
	}

	v, err := os.Open(filename)
	if err != nil {
		return &kubernetesManifestSpec{}, err
	}

	defer v.Close()

	content, err := io.ReadAll(v)
	if err != nil {
		return &kubernetesManifestSpec{}, err
	}

	err = goyaml.Unmarshal(content, &kubernetesData)

	if err != nil {
		return nil, err
	}

	return &kubernetesData, nil
}
