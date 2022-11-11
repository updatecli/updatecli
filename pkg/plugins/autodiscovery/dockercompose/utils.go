package dockercompose

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	goyaml "gopkg.in/yaml.v3"
)

// searchDockerComposeFiles will look, recursively, for every files named Chart.yaml from a root directory.
func searchDockerComposeFiles(rootDir string, files []string) ([]string, error) {

	dockerComposeFiles := []string{}

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
				dockerComposeFiles = append(dockerComposeFiles, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	logrus.Debugf("%d potential Docker compose file(s) found", len(dockerComposeFiles))

	return dockerComposeFiles, nil
}

// getDockerComposeData reads a Chart.yaml for information that could be automated
func getDockerComposeData(filename string) (*dockerComposeSpec, error) {

	var spec dockerComposeSpec

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

	err = goyaml.Unmarshal(content, &spec)

	if err != nil {
		return nil, err
	}

	if len(spec.Services) == 0 {
		return nil, nil
	}

	for _, service := range spec.Services {
		logrus.Debugf("Image: %q\n", service.Image)
	}

	return &spec, nil
}
