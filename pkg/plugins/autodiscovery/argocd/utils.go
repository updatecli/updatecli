package argocd

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	goyaml "go.yaml.in/yaml/v4"
)

// searchArgoCDFiles will look, recursively, for every YAML files from a root directory.
func searchArgoCDFiles(rootDir string, files []string) ([]string, error) {
	manifestFiles := []string{}

	logrus.Debugf("Looking for ArgoCD manifests in %q", rootDir)

	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
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
func readManifest(filename string) (map[int]*ArgoCDApplicationSpec, error) {

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

	loader, err := goyaml.NewLoader(bytes.NewReader(content), goyaml.V4)
	if err != nil {
		return nil, fmt.Errorf("creating yaml loader: %w", err)
	}

	docNum := 1
	result := make(map[int]*ArgoCDApplicationSpec)
	for {
		data := ArgoCDApplicationSpec{}

		err := loader.Load(&data)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("parsing yaml file %q: %w", filename, err)
		}

		// Check if the document contains a source definition that we can use to generate an Updatecli manifest
		if !data.Spec.Source.IsZero() ||
			!data.Spec.Template.Spec.Source.IsZero() ||
			len(data.Spec.Sources) > 0 {
			result[docNum-1] = &data
		}

		docNum++
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result, nil
}
