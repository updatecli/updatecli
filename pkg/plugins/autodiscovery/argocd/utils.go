package argocd

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	goyaml "gopkg.in/yaml.v3"
)

// searchArgoCDFiles will look, recursively, for every YAML files from a root directory.
func searchArgoCDFiles(rootDir string, files []string) ([]string, error) {

	manifestFiles := []string{}

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

				// First try to see if our file contains a HelmRelease definition
				data, err := loadApplicationData(path)
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
	for _, foundFile := range manifestFiles {
		fileName := filepath.Base(filepath.Dir(foundFile))
		logrus.Debugf("    * %q", fileName)
	}

	return manifestFiles, nil
}

// loadApplicationData reads a Chart.yaml for information that could be automated
func loadApplicationData(filename string) (*ArgoCDApplicationSpec, error) {

	var data ArgoCDApplicationSpec

	fileName := filepath.Base(filepath.Dir(filename))

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

	logrus.Debugf("ArgoCD manifest: %q\n", fileName)
	logrus.Debugf("Helm Chart Name: %q\n", data.Spec.Source.Chart)
	logrus.Debugf("Helm Repository URL: %q\n", data.Spec.Source.RepoURL)
	logrus.Debugf("Version: %q\n", data.Spec.Source.TargetRevision)

	return &data, nil
}
