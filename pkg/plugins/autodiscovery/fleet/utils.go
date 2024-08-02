package fleet

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	goyaml "gopkg.in/yaml.v3"
)

// searchFleetBundleFiles will look, recursively, for every files named Chart.yaml from a root directory.
func searchFleetBundleFiles(rootDir string, files []string) ([]string, error) {

	fleetBundleFiles := []string{}

	logrus.Debugf("Looking for Fleet bundle(s) in %q", rootDir)

	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		for _, f := range files {
			if info.Name() == f {
				fleetBundleFiles = append(fleetBundleFiles, path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	logrus.Debugf("%d Fleet bundle(s) found", len(fleetBundleFiles))
	for _, foundFile := range fleetBundleFiles {
		fleetBundleName := filepath.Base(filepath.Dir(foundFile))
		logrus.Debugf("    * %q", fleetBundleName)
	}

	return fleetBundleFiles, nil
}

// getFleetBundleData reads a Chart.yaml for information that could be automated
func getFleetBundleData(filename string) (*fleetMetadata, error) {

	var fleetData fleetMetadata

	fleetBundleName := filepath.Base(filepath.Dir(filename))
	logrus.Debugf("Fleet bundle %q found in %q", fleetBundleName, filepath.Dir(filename))

	if _, err := os.Stat(filename); err != nil {
		return &fleetMetadata{}, err
	}

	v, err := os.Open(filename)
	if err != nil {
		return &fleetMetadata{}, err
	}

	defer v.Close()

	content, err := io.ReadAll(v)
	if err != nil {
		return &fleetMetadata{}, err
	}

	err = goyaml.Unmarshal(content, &fleetData)

	if err != nil {
		return nil, err
	}

	logrus.Debugf("Fleet Bundle: %q\n", fleetBundleName)
	logrus.Debugf("Helm Chart Name: %q\n", fleetData.Helm.Chart)
	logrus.Debugf("Helm Repository URL: %q\n", fleetData.Helm.Repo)
	logrus.Debugf("Version: %q\n", fleetData.Helm.Version)

	return &fleetData, nil
}
