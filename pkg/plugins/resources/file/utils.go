package file

import (
	"net/http"
	"strings"
)

// isBinaryContent returns true if the content appears to be binary data.
// It uses net/http.DetectContentType which inspects up to the first 512 bytes
// and returns a MIME type — any type that is not "text/*" is treated as binary.
func isBinaryContent(content string) bool {
	contentType := http.DetectContentType([]byte(content))
	return !strings.HasPrefix(contentType, "text/")
}
