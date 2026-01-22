package argocd

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	goyaml "go.yaml.in/yaml/v3"
)

// searchArgoCDFiles will look, recursively, for every YAML files from a root directory.
func searchArgoCDFiles(rootDir string, files []string) ([]string, error) {
	manifestFiles := []string{}

	logrus.Debugf("Looking for ArgoCD manifests in %q", rootDir)

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

				if !match {
					continue
				}

				// First try to see if our file contains a Helm release definition
				data, err := readManifest(path)
				if err != nil {
					logrus.Debugf("Failed loading file %s as ArgoCD manifest: %s", path, err)
					continue
				}

				if data != nil {
					manifestFiles = append(manifestFiles, path)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	logrus.Debugf("%d Argocd manifest(s) found", len(manifestFiles))
	for _, manifestFile := range manifestFiles {
		manifestFile, err = filepath.Rel(rootDir, manifestFile)
		if err == nil {
			logrus.Debugf("    * %q", manifestFile)
		}
	}

	return manifestFiles, nil
}

// readManifest reads a Chart.yaml for information that could be automated
func readManifest(filename string) (*ArgoCDApplicationSpec, error) {
	var data ArgoCDApplicationSpec

	if _, err := os.Stat(filename); err != nil {
		return nil, err
	}

	v, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer v.Close()

	content, err := io.ReadAll(v)
	if err != nil {
		return nil, err
	}

	err = goyaml.Unmarshal(content, &data)
	if err != nil {
		return nil, err
	}

	if !data.Spec.Source.IsZero() {
		return &data, nil
	}

	for _, source := range data.Spec.Sources {
		if !source.IsZero() {
			return &data, nil
		}
	}

	return nil, nil
}
