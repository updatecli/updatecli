package helm

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	goyaml "gopkg.in/yaml.v3"
)

// searchChartMetadataFiles will look, recursively, for every files named Chart.yaml from a root directory.
func searchChartMetadataFiles(rootDir string, files []string) ([]string, error) {

	metadataFiles := []string{}

	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		for _, f := range files {
			if info.Name() == f {
				metadataFiles = append(metadataFiles, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	logrus.Debugf("%d chart(s) found", len(metadataFiles))
	for _, foundFile := range metadataFiles {
		chartName := filepath.Base(filepath.Dir(foundFile))
		logrus.Debugf("    * %q", chartName)
	}

	return metadataFiles, nil
}

// getChartMetadata reads a Chart.yaml for information that could be automated
func getChartMetadata(filename string) (*chartMetadata, error) {

	var chart chartMetadata

	chartName := filepath.Base(filepath.Dir(filename))
	logrus.Infof("Chart %q found in %q", chartName, filepath.Dir(filename))

	if _, err := os.Stat(filename); err != nil {
		return &chartMetadata{}, err
	}

	v, err := os.Open(filename)
	if err != nil {
		return &chartMetadata{}, err
	}

	defer v.Close()

	content, err := ioutil.ReadAll(v)
	if err != nil {
		return &chartMetadata{}, err
	}

	err = goyaml.Unmarshal(content, &chart)

	if err != nil {
		return nil, err
	}

	if len(chart.Dependencies) == 0 {
		return &chartMetadata{}, nil
	}

	logrus.Debugf("Chart: %q\n", chartName)
	for _, value := range chart.Dependencies {
		logrus.Debugf("Name: %q\n", value.Name)
		logrus.Debugf("URL: %q\n", value.Repository)
		logrus.Debugf("Version: %q\n", value.Version)
	}

	return &chart, nil
}
