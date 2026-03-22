package file

import (
	"net/http"
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

// isBinaryContent returns true if the content appears to be binary data.
// It uses net/http.DetectContentType which inspects up to the first 512 bytes
// and returns a MIME type — any type that is not "text/*" is treated as binary.
func isBinaryContent(content string) bool {
	contentType := http.DetectContentType([]byte(content))
	return !strings.HasPrefix(contentType, "text/")
}
