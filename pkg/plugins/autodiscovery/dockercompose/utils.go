package dockercompose

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/sirupsen/logrus"
	goyaml "gopkg.in/yaml.v3"
)

// searchDockerComposeFiles will look, recursively, for every files named Chart.yaml from a root directory.
func searchDockerComposeFiles(rootDir string, filePatterns []string) ([]string, error) {
	dockerComposeFiles := []string{}

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if !d.IsDir() {
			for _, f := range filePatterns {
				match, err := filepath.Match(f, d.Name())
				if err != nil {
					logrus.Errorln(err)
					continue
				}
				if match {
					dockerComposeFiles = append(dockerComposeFiles, path)
				}
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

// getDockerComposeSpecFromFile reads a Chart.yaml for information that could be automated
func getDockerComposeSpecFromFile(filename string) (dockercomposeServicesList, error) {
	type dockerComposeSpec struct {
		Services map[string]dockerComposeServiceSpec
	}
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

	result := make([]dockerComposeService, 0, len(spec.Services))

	// Generate the outgoing collection by sorting sub-keys of the top-level `services` key
	services := make([]string, 0, len(spec.Services))
	for svc := range spec.Services {
		services = append(services, svc)
	}
	sort.Strings(services)

	for _, svcName := range services {
		logrus.Debugf("Found service definition: %v for the docker compose service %q\n", spec.Services[svcName], svcName)
		result = append(result, dockerComposeService{
			Name: svcName,
			Spec: spec.Services[svcName],
		})
	}

	return result, nil
}
