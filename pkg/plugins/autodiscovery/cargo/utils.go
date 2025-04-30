package cargo

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"

	sv "github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils/dasel"
)

func isCargoUpgradeAvailable() bool {
	return exec.Command("cargo", "upgrade", "--version").Run() == nil
}

func isCargoAvailable() bool {
	return exec.Command("cargo", "--version").Run() == nil
}

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
				return fs.SkipDir
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
			if _, err := fc.DaselNode.Query(fmt.Sprintf(".%s.%s.git", dependencyType, pkg)); err == nil {
				// TODO: handle Git dependencies
				continue
			}
			// It could have been defined using a path to a local directory
			if _, err = fc.DaselNode.Query(fmt.Sprintf(".%s.%s.path", dependencyType, pkg)); err == nil {
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

func getCrateMetadata(rootDir string) (crateMetadata, error) {
	manifestPath := fmt.Sprintf("%s/Cargo.toml", rootDir)
	lockPath := fmt.Sprintf("%s/Cargo.lock", rootDir)
	if _, err := os.Stat(lockPath); err != nil {
		lockPath = ""
	}

	crate := crateMetadata{
		CargoFile:     manifestPath,
		CargoLockFile: lockPath,
	}

	tomlFile := dasel.FileContent{
		DataType:         "toml",
		FilePath:         manifestPath,
		ContentRetriever: &text.Text{},
	}

	err := tomlFile.Read("")

	if err != nil {
		return crate, err
	}

	workspaceMembers := []crateMetadata{}
	if members, err := tomlFile.MultipleQuery("workspace.members.[*]"); err == nil {
		for _, member := range members {
			memberPath := fmt.Sprintf("%s/%s", rootDir, member)
			if matches, err := filepath.Glob(memberPath); err == nil {
				for _, match := range matches {
					if workspaceMember, err := getCrateMetadata(match); err == nil {
						workspaceMember.CargoLockFile = lockPath
						workspaceMembers = append(workspaceMembers, workspaceMember)
					}
				}
			}
		}
		if len(workspaceMembers) > 0 {
			crate.Workspace = true
		}
	}
	name, err := tomlFile.Query("package.name")

	if err != nil && !crate.Workspace {
		// No package, and no workspace members
		return crate, err
	} else {
		crate.Name = name
	}

	crate.WorkspaceMembers = workspaceMembers
	crate.Dependencies, _ = getDependencies(&tomlFile, "dependencies")
	crate.DevDependencies, _ = getDependencies(&tomlFile, "dev-dependencies")
	crate.WorkspaceDependencies, _ = getDependencies(&tomlFile, "workspace.dependencies")
	crate.WorkspaceDevDependencies, _ = getDependencies(&tomlFile, "workspace.dev-dependencies")

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

	return crate, nil
}

func isStrictSemver(version string) bool {
	if _, err := sv.StrictNewVersion(version); err != nil {
		return false
	}
	return true
}
