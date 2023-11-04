package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

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
func FindFilesMatchingPathPattern(filePathPattern string) ([]string, error) {

	if strings.HasPrefix(filePathPattern, "https://") ||
		strings.HasPrefix(filePathPattern, "http://") {

		logrus.Debugln("file path is an URL, skipping file search")

		return []string{filePathPattern}, nil
	}

	filePathPattern = strings.TrimPrefix(filePathPattern, "file://")

	results := []string{}

	rootDir := "."
	if filepath.IsAbs(filePathPattern) {
		rootDir = filepath.Dir(filePathPattern)
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logrus.Errorf("\n%s File %s: %s\n", result.FAILURE, path, err)
			return fmt.Errorf("unable to walk %q: %s", path, err)
		}
		if info.Mode().IsRegular() {
			match, err := filepath.Match(filePathPattern, path)
			if err != nil {
				return fmt.Errorf("unable to match %q: %s", filePathPattern, err)
			}
			if match {
				results = append(results, path)
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
