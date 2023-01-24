package cargo

import (
	"fmt"
	"io/fs"
	"path/filepath"

	sv "github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/dasel"
)

// findCargoFiles search, recursively, for every files named Cargo.toml from a root directory.
func findCargoFiles(rootDir string, validFiles []string) ([]string, error) {
	cargoFiles := []string{}

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
				cargoFiles = append(cargoFiles, path)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	logrus.Debugf("%d cargo files(s) found", len(cargoFiles))
	for _, foundFile := range cargoFiles {
		cargoFile := filepath.Base(filepath.Dir(foundFile))
		logrus.Debugf("    * %q", cargoFile)
	}
	return cargoFiles, nil
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
		version, err := fc.DaselNode.Query(fmt.Sprintf(".%s.%s.version", dependencyType, pkg))
		if err != nil {
			// Cargo dependency has not been defined using a version
			// It could have been defined using a git repository
			_, err := fc.DaselNode.Query(fmt.Sprintf(".%s.%s.git", dependencyType, pkg))
			if err == nil {
				// TODO: Handle git dependencies
				continue
			}
			// It could have been defined using a path to a local directory
			_, err = fc.DaselNode.Query(fmt.Sprintf(".%s.%s.path", dependencyType, pkg))
			if err == nil {
				// TODO: Handle Path dependencies
				continue
			}

			version, err := fc.DaselNode.Query(fmt.Sprintf(".%s.%s", dependencyType, pkg))
			if err != nil {
				continue
			}
			// Ensure version is semver compliant
			v := version.String()
			if _, err = sv.NewVersion(v); err != nil {
				continue
			}
			cd.Version = v
			cd.Inlined = true
		} else {
			cd.Version = version.String()
		}
		registry, _ := fc.DaselNode.Query(fmt.Sprintf(".%s.%s.registry", dependencyType, pkg))
		if err == nil && registry != nil {
			cd.Registry = registry.String()
		}
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

	err := tomlFile.Read("")

	if err != nil {
		return &crateMetadata{}, err
	}

	name, err := tomlFile.Query("package.name")

	if err != nil {
		return &crateMetadata{}, err
	}

	crate.Name = name
	crate.Dependencies, _ = getDependencies(&tomlFile, "dependencies")
	crate.DevDependencies, _ = getDependencies(&tomlFile, "dev-dependencies")

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
