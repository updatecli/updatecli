package helm

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	goyaml "go.yaml.in/yaml/v3"
)

// searchChartFiles search, recursively, for every files named Chart.yaml from a root directory.
func searchChartFiles(rootDir string, files []string) ([]string, error) {
	metadataFiles := []string{}

	logrus.Debugf("Looking for Helm charts in %q", rootDir)

	// To do switch to WalkDir which is more efficient, introduced in 1.16
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		for _, f := range files {
			if info.Name() == f {
				if isChartRootDirectory(path) {
					metadataFiles = append(metadataFiles, path)
				}
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

// isChartRootDirectory checks that file provided by argument locates at the root of a Chart directory
func isChartRootDirectory(path string) bool {
	for _, chartFile := range ChartValidFiles {
		// If browse file is Chart.yaml or Chart.yml then Updatecli assumes that it's a Chart root directory
		if chartFile == filepath.Base(path) {
			return true
		}

		if _, err := os.Stat(filepath.Join(filepath.Dir(path), chartFile)); err == nil {
			return true
		}
	}
	return false
}

// getChartMetadata reads a Chart.yaml for information.
func getChartMetadata(filename string) (*chartMetadata, error) {
	var chart chartMetadata

	chartName := filepath.Base(filepath.Dir(filename))
	logrus.Debugf("Chart %q found in %q", chartName, filepath.Dir(filename))

	if _, err := os.Stat(filename); err != nil {
		return &chartMetadata{}, err
	}

	v, err := os.Open(filename)
	if err != nil {
		return &chartMetadata{}, err
	}

	defer v.Close()

	content, err := io.ReadAll(v)
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

// getValuesFileContent reads a values.yaml for information.
func getValuesFileContent(filename string) (*valuesContent, error) {
	var values valuesContent

	chartName := filepath.Base(filepath.Dir(filename))
	logrus.Debugf("Chart values file %q found in %q", chartName, filepath.Dir(filename))

	if _, err := os.Stat(filename); err != nil {
		return &valuesContent{}, err
	}

	v, err := os.Open(filename)
	if err != nil {
		return &valuesContent{}, err
	}

	defer v.Close()

	content, err := io.ReadAll(v)
	if err != nil {
		return &valuesContent{}, err
	}

	err = goyaml.Unmarshal(content, &values)
	if err != nil {
		return nil, err
	}

	if len(values.Images) == 0 && values.Image.Repository == "" {
		return &valuesContent{}, nil
	}

	if len(values.Images) > 0 {
		logrus.Debugf("Images found for chart %q\n", chartName)
		for id, value := range values.Images {
			logrus.Debugf("Image id %q found\n", id)
			logrus.Debugf("\tName: %q\n", value.Repository)
			logrus.Debugf("\tURL: %q\n", value.Tag)
		}
	}

	if values.Image.Repository != "" {
		logrus.Debugf("Image Repository: %q\n", values.Image.Repository)
		logrus.Debugf("Image Tag: %q\n", values.Image.Tag)
	}

	return &values, nil
}
