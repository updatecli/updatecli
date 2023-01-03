package npm

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

const (
	PackageJsonFile string = "package.json"
)

// searchPackageJsonFiles looks, recursively, for every files named package.json from a root directory.
func searchPackageJsonFiles(rootDir string) ([]string, error) {

	foundFiles := []string{}

	// To do switch to WalkDir which is more efficient, introduced in 1.16
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		// Updatecli should ignore all package.json from the directory named "node_modules"
		// as they are automatically installed by npm
		if info.IsDir() && info.Name() == "node_modules" {
			return filepath.SkipDir
		}

		if info.Name() == PackageJsonFile {
			foundFiles = append(foundFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return foundFiles, nil
}

// isVersionConstraintSpecified checks if the version is a version constraint that should be handle by npm
// and not updatecli
// https://docs.npmjs.com/cli/v6/configuring-npm/package-json#dependencies
func isVersionConstraintSpecified(packageName, packageVersion string, strictSemver bool) bool {
	// version set to an empty string is equivalent to *

	if packageVersion == "" || packageVersion == "latest" || packageVersion == "*" {
		logrus.Debugf("Ignoring dependency %q. It contains a version constraint %q handled by NPM", packageName, packageVersion)
		logrus.Debugln("You probably want to adopt a better versioning strategy")
		return true
	}

	// Check version start with
	for _, toIgnorePrefix := range []string{">", "<", "~", "^", "*"} {
		if strings.HasPrefix(packageVersion, toIgnorePrefix) {
			logrus.Debugf("Ignoring dependency %q. It contains a version constraint %q handled by NPM", packageName, packageVersion)
			return true
		}
	}

	// Ignore http urls paths dependencies
	// https://docs.npmjs.com/cli/v6/configuring-npm/package-json#git-urls-as-dependencies
	for _, uriPrefix := range []string{"https://", "http://", "ftp://", "file://"} {
		if strings.HasPrefix(packageVersion, uriPrefix) {
			logrus.Debugf("Ignoring dependency %q. Updating URL is not supported at this time. Feel free to reach out to suggest an update scenario", packageName)
			return true
		}
	}

	// Ignore git urls paths dependencies
	// https://docs.npmjs.com/cli/v6/configuring-npm/package-json#git-urls-as-dependencies
	for _, gitPrefix := range []string{"git://", "git+ssh://", "git+https://", "git+file://"} {
		if strings.HasPrefix(packageVersion, gitPrefix) {
			logrus.Debugf("Ignoring dependency %q. Updating Git URL is not supported at this time. Feel free to reach out to suggest an update scenario", packageName)
			return true
		}
	}

	// Ignore local paths dependencies
	// https://docs.npmjs.com/cli/v6/configuring-npm/package-json#local-paths
	for _, localPath := range []string{"../", "./", "~/", "/"} {
		if strings.HasPrefix(packageVersion, localPath) {
			logrus.Debugf("Ignoring dependency %q. Updating local path is not supported at this time. Feel free to reach out to suggest an update scenario", packageName)
			return true
		}
	}

	switch strictSemver {
	case true:
		_, err := semver.StrictNewVersion(packageVersion)
		if err != nil {
			logrus.Debugln(err)
			logrus.Debugf("NPM package version %s detected. Updatecli only updates version following strict semantic versioning", packageVersion)
			return true
		}
	case false:
		_, err := semver.NewVersion(packageVersion)
		if err != nil {
			logrus.Debugln(err)
			logrus.Debugf("NPM package version %s detected. Updatecli only updates version following semantic versioning", packageVersion)
			return true
		}
	}

	return false
}

func isNpmInstalled() bool {
	cmd := exec.Command("npm", "--version")
	err := cmd.Run()
	return err == nil
}

func isYarnInstalled() bool {
	cmd := exec.Command("yarn", "--version")
	err := cmd.Run()
	return err == nil
}

func isLockFileDetected(lockfile string) bool {
	_, err := os.Stat(lockfile)
	return err == nil
}
