package yaml

import (
	"path/filepath"
	"strings"
)

// joinPathwithworkingDirectoryPath To merge File path with current workingDir, unless file is an HTTP URL
func joinPathWithWorkingDirectoryPath(fileName, workingDir string) string {
	if workingDir == "" ||
		filepath.IsAbs(fileName) ||
		strings.HasPrefix(fileName, "https://") ||
		strings.HasPrefix(fileName, "http://") {
		return fileName
	}

	return filepath.Join(workingDir, fileName)
}
