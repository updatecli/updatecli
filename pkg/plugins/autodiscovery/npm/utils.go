package npm

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
)

const (
	PackageJsonFile         string = "package.json"
	latestVersionIdentifier string = "latest"
)

var (
	// semverRegex matches valid semantic versioning such as 1.0.0, 1.0, 1
	semverRegex = regexp.MustCompile(`\d+(\.\d+)?(\.\d+)?`)
)

// searchPackageJsonFiles looks, recursively, for every files named package.json from a root directory.
func searchPackageJsonFiles(rootDir string) ([]string, error) {

	foundFiles := []string{}

	logrus.Debugf("Looking for package.json files in %q", rootDir)

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

// isVersionConstraintSupported checks if the version is a version constraint that should be handle by npm
// and not updatecli
// https://docs.npmjs.com/cli/v6/configuring-npm/package-json#dependencies
func isVersionConstraintSupported(packageName, packageVersion string) bool {
	// Ignore http urls paths dependencies
	// https://docs.npmjs.com/cli/v6/configuring-npm/package-json#git-urls-as-dependencies

	uriPrefixes := []string{
		"https://",
		"http://",
		"ftp://",
		"file://",
		"git://",
		"git+ssh://",
		"git+https://",
		"git+file://",
	}

	for _, uriPrefix := range uriPrefixes {
		if strings.HasPrefix(packageVersion, uriPrefix) {
			logrus.Debugf("Ignoring dependency %q. Updating uri type %q is not supported at this time. Feel free to reach out to suggest an update scenario", packageName, uriPrefix)
			return false
		}
	}

	// Ignore local paths dependencies
	// https://docs.npmjs.com/cli/v6/configuring-npm/package-json#local-paths
	for _, localPath := range []string{"../", "./", "~/", "/"} {
		if strings.HasPrefix(packageVersion, localPath) {
			logrus.Debugf("Ignoring dependency %q. Updating local path is not supported at this time. Feel free to reach out to suggest an update scenario", packageName)
			return false
		}
	}

	if packageVersion == "" || packageVersion == latestVersionIdentifier || packageVersion == "*" {
		logrus.Debugf("Ignoring dependency %q. It contains a version constraint %q handled by NPM", packageName, packageVersion)
		logrus.Debugln("You probably want to adopt a better versioning strategy")
		return true
	}

	_, err := semver.NewConstraint(packageVersion)
	if err != nil {
		logrus.Debugln(err)
		logrus.Debugf("Semantic versioning constraint %q not supported for package %q", packageVersion, packageName)
		return false
	}

	return true
}

// isVersionConstraintSpecified checks if the version is a version constraint that should be handle by npm
// and not updatecli
// https://docs.npmjs.com/cli/v6/configuring-npm/package-json#dependencies
func isVersionConstraintSpecified(packageName, packageVersion string) bool {
	// version set to an empty string is equivalent to *

	if packageVersion == "" || packageVersion == latestVersionIdentifier || packageVersion == "*" {
		logrus.Debugf("Ignoring dependency %q. It contains a version constraint %q handled by NPM", packageName, packageVersion)
		logrus.Debugln("You probably want to adopt a better versioning strategy")
		return true
	}

	// Check version start with
	for _, toIgnorePrefix := range []string{">", "<", "~", "^", "*"} {
		if strings.HasPrefix(packageVersion, toIgnorePrefix) {
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

	_, err := semver.StrictNewVersion(packageVersion)
	if err != nil {
		logrus.Debugln(err)
		logrus.Debugf("None strict semantic version detected %s for package %q", packageVersion, packageName)
		return true
	}

	return false
}

func isNpmInstalled() bool {
	cmd := exec.Command("npm", "--version")
	err := cmd.Run()
	return err == nil
}

// Since npm version 8, npm support updating yarn.lock file when it detects it
// isNpmSupportYarnUpdate checks if the current npm version can update the yarn file which is the preferred approach as it supports a dryrun mode
func isNpmSupportYarnUpdate() bool {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("npm", "--version")

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		logrus.Debugf("failed while running npm command - %s", stderr.String())
		return false
	}

	currentVersion, err := semver.NewVersion(strings.TrimSuffix(stdout.String(), "\n"))
	if err != nil {
		logrus.Debugf("failed verifying current npm version - %s\n%s", err, stdout.String())
		return false
	}

	c, err := semver.NewConstraint(">= 8.0.0")
	if err != nil {
		logrus.Debugln(err)
	}

	return c.Check(currentVersion)
}

func isYarnInstalled() bool {
	cmd := exec.Command("yarn", "--version")
	err := cmd.Run()
	return err == nil
}

func isPnpmInstalled() bool {
	cmd := exec.Command("pnpm", "--version")
	err := cmd.Run()
	return err == nil
}

func isLockFileDetected(lockfile string) bool {
	_, err := os.Stat(lockfile)
	return err == nil
}

func getTargetCommand(cmd, dependencyName string) string {
	dryRunVariable := "$DRY_RUN"
	if runtime.GOOS == "windows" {
		dryRunVariable = "$env:DRY_RUN"
	}

	switch cmd {
	case "npm":
		return fmt.Sprintf("npm install --package-lock-only --dry-run=%s %s@{{ source %q }}", dryRunVariable, dependencyName, "npm")
	case "yarn":
		if isNpmSupportYarnUpdate() {
			return fmt.Sprintf("npm install --package-lock-only --dry-run=%s %s@{{ source %q }}", dryRunVariable, dependencyName, "npm")
		}
		logrus.Warningf("In the current state, yarn package update do not support dry-run mode")
		return fmt.Sprintf("yarn add --mode update-lockfile %s@{{ source %q }}", dependencyName, "npm")
	case "pnpm":
		logrus.Warningf("In the current state, pnpm package update does not support dry-run mode")
		return fmt.Sprintf("pnpm add --lockfile-only %s@{{ source %q }}", dependencyName, "npm")
	}

	return "false"
}

// convertSemverVersionConstraintToVersion tries to extract a valid semantic version from a version constraint
// If it fails to extract a valid semantic version, it returns an error
// If the version constraint is "latest", it returns an empty string and no error
// If the version constraint is a strict semantic version, it returns it as is
// If the version constraint is a valid semantic version constraint, it extracts the first valid semantic version from it
func convertSemverVersionConstraintToVersion(versionConstraint string) (string, error) {
	// If the version constraint is already a strict version, return it as is

	if versionConstraint == latestVersionIdentifier {
		return "", nil
	}

	if _, err := semver.NewConstraint(versionConstraint); err != nil {
		return "", fmt.Errorf("parsing version constraint %q: %s", versionConstraint, err)
	}

	match := semverRegex.FindString(versionConstraint)
	if match == "" {
		return "", fmt.Errorf("no valid version found in constraint: %s", versionConstraint)
	}

	version, err := semver.NewVersion(match)
	if err != nil {
		return "", fmt.Errorf("parsing version from constraint %q: %s", versionConstraint, err)
	}

	return version.String(), nil
}
