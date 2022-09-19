package helmfile

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	goyaml "gopkg.in/yaml.v3"
)

// searchHelmfileFiles will look, recursively, for every files named Chart.yaml from a root directory.
func searchHelmfileFiles(rootDir string, files []string) ([]string, error) {

	helmfiles := []string{}

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

// getHelmfileMetadata reads a Chart.yaml for information that could be automated
func getHelmfileMetadata(filename string) (*helmfileMetadata, error) {

	var helmfile helmfileMetadata

	if _, err := os.Stat(filename); err != nil {
		return &helmfileMetadata{}, err
	}

	v, err := os.Open(filename)
	if err != nil {
		return &helmfileMetadata{}, err
	}

	defer v.Close()

	content, err := ioutil.ReadAll(v)
	if err != nil {
		return &helmfileMetadata{}, err
	}

	err = goyaml.Unmarshal(content, &helmfile)

	if err != nil {
		return nil, err
	}

	if len(helmfile.Releases) == 0 {
		return &helmfileMetadata{}, nil
	}

	for _, value := range helmfile.Releases {
		logrus.Debugf("Name: %q\n", value.Name)
		logrus.Debugf("Version: %q\n", value.Version)
	}

	return &helmfile, nil
}

func getReleaseRepositoryUrl(repositories []repository, release release) (name, url string) {
	for i := range repositories {
		if strings.HasPrefix(release.Chart, repositories[i].Name+"/") {
			return strings.TrimPrefix(release.Chart, repositories[i].Name+"/"), repositories[i].URL
		}
	}
	return "", ""
}
