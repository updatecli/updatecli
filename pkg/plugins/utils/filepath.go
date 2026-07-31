package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// SanitizeFilePathWithWorkingDirectory resolves filePath against workingDir and
// guarantees that the returned path stays within workingDir. It is the single
// containment primitive shared by every file based resource (file, yaml, xml,
// toml, toolversions, json, csv) so the invariant cannot drift between them.
//
// Behavior:
//   - http:// and https:// URLs are returned unchanged (remote reads).
//   - When workingDir is empty there is no containment boundary — updatecli is
//     running without an SCM checkout — so the path is returned unchanged to
//     preserve the historical local run behavior.
//   - When workingDir is set, an absolute filePath is rejected and any relative
//     filePath that would escape workingDir through ".." traversal is rejected.
//
// Rejecting instead of silently clamping is deliberate: a path that resolves
// outside the working directory almost always means an upstream value (e.g. a
// {{ source }} output) has been injected into a target/source path, and the
// pipeline must fail rather than read from or write to an attacker chosen
// location. See GHSA-hj4x-hm4v-7wpw.
func SanitizeFilePathWithWorkingDirectory(filePath, workingDir string) (string, error) {
	if strings.HasPrefix(filePath, "https://") ||
		strings.HasPrefix(filePath, "http://") {
		return filePath, nil
	}

	// Without a working directory there is no boundary to contain the path to,
	// so we keep the caller provided path as is (local run, no SCM checkout).
	if workingDir == "" {
		return filePath, nil
	}

	if filepath.IsAbs(filePath) {
		return "", fmt.Errorf(
			"absolute path %q is not allowed: files must stay within the working directory %q",
			filePath, workingDir)
	}

	joinedPath := filepath.Join(workingDir, filePath)

	relPath, err := filepath.Rel(workingDir, joinedPath)
	if err != nil {
		return "", fmt.Errorf(
			"unable to verify that path %q stays within the working directory %q: %w",
			filePath, workingDir, err)
	}

	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf(
			"path %q escapes the working directory %q",
			filePath, workingDir)
	}

	return joinedPath, nil
}

// JoinPathwithworkingDirectoryPath To merge File path with current workingDir, unless file is an HTTP URL
func JoinFilePathWithWorkingDirectoryPath(filePath, workingDir string) string {

	currentWorkingDirectory, err := os.Getwd()
	if err != nil {
		logrus.Debugln("fail getting current working directory")
		return filePath
	}

	if currentWorkingDirectory == workingDir {
		return filePath
	}

	if workingDir == "" ||
		filepath.IsAbs(filePath) ||
		strings.HasPrefix(filePath, "https://") ||
		strings.HasPrefix(filePath, "http://") {
		return filePath
	}

	return filepath.Join(workingDir, filePath)
}

// FindFilesMatchingPathPattern returns a list of files matching a file pattern
func FindFilesMatchingPathPattern(rootDir, filePathPattern string) ([]string, error) {

	if strings.HasPrefix(filePathPattern, "https://") ||
		strings.HasPrefix(filePathPattern, "http://") {

		logrus.Debugln("file path is an URL, skipping file search")

		return []string{filePathPattern}, nil
	}

	filePathPattern = strings.TrimPrefix(filePathPattern, "file://")

	results := []string{}

	if rootDir == "" {
		var err error
		rootDir, err = os.Getwd()
		if err != nil {
			return []string{}, fmt.Errorf("unable to get current working directory: %s", err)
		}
	}

	if filepath.IsAbs(filePathPattern) {
		rootDir = filepath.Dir(filePathPattern)
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logrus.Errorf("\n%s File %s: %s\n", result.FAILURE, path, err)
			return fmt.Errorf("unable to walk %q: %s", path, err)
		}

		if info.Mode().IsRegular() {
			relPath, err := filepath.Rel(rootDir, path)
			if err != nil {
				return fmt.Errorf("unable to get relative path of %q: %s", path, err)
			}

			match, err := filepath.Match(filePathPattern, relPath)
			if err != nil {
				return fmt.Errorf("unable to match %q: %s", filePathPattern, err)
			}
			if match {
				results = append(results, relPath)
			}
		}
		return nil
	})

	if err != nil {
		return []string{}, fmt.Errorf("unable to walk %q: %s", rootDir, err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("unable to find files matching %q", filePathPattern)
	}

	logrus.Debugf("Found %d files matching %q", len(results), filePathPattern)

	return results, nil

}
