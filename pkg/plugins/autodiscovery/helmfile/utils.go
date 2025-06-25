package helmfile

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	goyaml "go.yaml.in/yaml/v3"
)

// searchHelmfileFiles search, recursively, for every Helmfile files starting from a root directory.
func searchHelmfileFiles(rootDir string, files []string) ([]string, error) {

	helmfiles := []string{}

	logrus.Debugf("Looking for Helmfile(s) in %q", rootDir)

	// To do switch to WalkDir which is more efficient, introduced in 1.16
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		for _, f := range files {
			match, err := filepath.Match(f, info.Name())
			if err != nil {
				logrus.Errorln(err)
				continue
			}
			if match {
				helmfiles = append(helmfiles, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	logrus.Debugf("%d potential helmfile(s) found", len(helmfiles))

	return helmfiles, nil
}

// getHelmfileMetadata loads file content from a Helmfile file.
func getHelmfileMetadata(filename string) (*helmfileMetadata, error) {

	var helmfile helmfileMetadata

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

	err = goyaml.Unmarshal(content, &helmfile)

	if err != nil {
		return nil, err
	}

	if len(helmfile.Releases) == 0 {
		return nil, nil
	}

	for _, value := range helmfile.Releases {
		logrus.Debugf("Name: %q\n", value.Name)
		logrus.Debugf("Version: %q\n", value.Version)
	}

	return &helmfile, nil
}
