package yaml

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

func parseKey(key string) []string {

	elements := []string{}
	element := ""
	escapedCharacter := false

	for i := range key {
		fmt.Printf("Current element: %q\n", element)
		switch string(key[i]) {
		case `\`:
			if !escapedCharacter {
				escapedCharacter = true
			}

		case `.`:
			if escapedCharacter {
				element = element + string(key[i])
				escapedCharacter = false
				continue
			}

			elements = append(elements, element)
			element = ""

		default:
			element = element + string(key[i])
		}
	}

	if len(element) > 0 {
		elements = append(elements, element)
	}

	return elements
}
