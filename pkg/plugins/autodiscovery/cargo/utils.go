package cargo

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/dasel"
)

// searchChartFiles search, recursively, for every files named Cargo.toml from a root directory.
func findCargoFiles(rootDir string, validFiles []string) ([]string, error) {
	manifestsFiles := []string{}

	err := filepath.WalkDir(rootDir, func(path string, di fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if di.IsDir() {
			return nil
		}

		for _, f := range validFiles {
			if di.Name() == f {
				manifestsFiles = append(manifestsFiles, path)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	logrus.Debugf("%d chart(s) found", len(manifestsFiles))
	for _, foundFile := range manifestsFiles {
		chartName := filepath.Base(filepath.Dir(foundFile))
		logrus.Debugf("    * %q", chartName)
	}
	return manifestsFiles, nil
}

func getDependencies(fc *dasel.FileContent, dependencyType string) ([]crateDependency, error) {
	var dependencies []crateDependency
	packages, err := fc.MultipleQuery(fmt.Sprintf(".%s.-", dependencyType))
	if err != nil {
		return dependencies, err
	}
	for _, pkg := range packages {
		cd := crateDependency{
			Name: pkg,
		}
		cd.Version, err = fc.Query(fmt.Sprintf(".%s.%s.version", dependencyType, pkg))
		if err != nil {
			cd.Version, err = fc.Query(fmt.Sprintf(".%s.%s", dependencyType, pkg))
			if err != nil {
				continue
			}
			cd.Inlined = true
		}
		cd.Registry, _ = fc.Query(fmt.Sprintf(".%s.%s.registry", dependencyType, pkg))
		dependencies = append(dependencies, cd)
	}
	return dependencies, nil
}

func getCrateMetadata(manifestPath string) (*crateMetadata, error) {

	var crate crateMetadata

	tomlFile := dasel.FileContent{
		DataType:         "toml",
		FilePath:         manifestPath,
		ContentRetriever: &text.Text{},
	}

	err := tomlFile.Read(".")

	if err != nil {
		return &crateMetadata{}, err
	}

	name, err := tomlFile.Query("package.name")

	if err != nil {
		return &crateMetadata{}, err
	}

	crate.Name = name
	crate.Dependencies, err = getDependencies(&tomlFile, "dependencies")
	if err != nil {
		return &crateMetadata{}, err
	}
	crate.DevDependencies, err = getDependencies(&tomlFile, "dev-dependencies")
	if err != nil {
		return &crateMetadata{}, err
	}

	logrus.Debugf("Crate: %q\n", name)
	logrus.Debugf("Dependencies")
	for _, value := range crate.Dependencies {
		logrus.Debugf("Name: %q\n", value.Name)
		logrus.Debugf("Registry: %q\n", value.Registry)
		logrus.Debugf("Version: %q\n", value.Version)
	}
	logrus.Debugf("Dev-Dependencies")
	for _, value := range crate.DevDependencies {
		logrus.Debugf("Name: %q\n", value.Name)
		logrus.Debugf("Registry: %q\n", value.Registry)
		logrus.Debugf("Version: %q\n", value.Version)
	}

	return &crate, nil
}
