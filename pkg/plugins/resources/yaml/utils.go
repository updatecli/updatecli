package yaml

import (
	"path/filepath"
	"strings"
)

// joinPathwithworkingDirectoryPath To merge File path with current workingDir, unless file is an HTTP URL
func joinPathWithWorkingDirectoryPath(filePath, workingDir string) string {
	if workingDir == "" ||
		filepath.IsAbs(filePath) ||
		strings.HasPrefix(filePath, "https://") ||
		strings.HasPrefix(filePath, "http://") {
		return filePath
	}

	return filepath.Join(workingDir, filePath)
}
