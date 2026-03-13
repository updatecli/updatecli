package file

import (
	"fmt"
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

// isBinaryContent checks if the given content appears to be binary data
// by looking for null bytes which are characteristic of binary files
func isBinaryContent(content string) bool {
	checkSize := len(content)
	if checkSize > 8192 {
		checkSize = 8192
	}

	for i := 0; i < checkSize; i++ {
		if content[i] == 0 {
			return true
		}
	}
	return false
}

// truncateBinaryContent returns a placeholder message for binary content
func truncateBinaryContent(content string) string {
	return fmt.Sprintf("[binary content, %d bytes]", len(content))
}
