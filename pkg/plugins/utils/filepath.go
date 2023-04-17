package utils

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
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
