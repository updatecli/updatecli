package ko

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	goyaml "go.yaml.in/yaml/v3"
)

// searchKosFiles will look, recursively, for every Ko files.
func searchKosFiles(rootDir string, files []string) ([]string, error) {
	logrus.Debugf("Looking for Ko files in %q", rootDir)

	koFiles := []string{}

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
					koFiles = append(koFiles, path)
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	logrus.Debugf("%d Ko file(s) found", len(koFiles))

	for _, foundFile := range koFiles {
		koFile := filepath.Base(filepath.Dir(foundFile))
		logrus.Debugf("    * %q", koFile)
	}

	return koFiles, nil
}

// getKoManifestData reads a Ko file for information that could be automatically updated.
func getKoManifestData(filename string) (*koSpec, error) {
	var data koSpec

	koFile := filepath.Base(filepath.Dir(filename))
	logrus.Debugf("Ko file %q found in %q", koFile, filepath.Dir(filename))

	if _, err := os.Stat(filename); err != nil {
		return &koSpec{}, err
	}

	v, err := os.Open(filename)
	if err != nil {
		return &koSpec{}, err
	}

	defer v.Close()

	content, err := io.ReadAll(v)
	if err != nil {
		return &koSpec{}, err
	}

	err = goyaml.Unmarshal(content, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}
