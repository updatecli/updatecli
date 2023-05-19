package yaml

import (
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
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

/*
sanitizeYamlPathKey is a helper function to migrate the depecrated yaml key
to the new syntax. We start by displaying a warning message, and the next step,
will be to return an error.
*/
func sanitizeYamlPathKey(key string) string {

	elements := []string{}
	tmpElements := strings.Split(key, `.`)

	quoted := false
	for i, elem := range tmpElements {
		switch quoted {
		case true:
			if strings.HasSuffix(elem, `\`) {
				elements = append(elements, strings.TrimSuffix(elem, `\`))
			} else {
				elements = append(elements, elem+`'`)
				quoted = false
			}

		case false:
			if strings.HasSuffix(elem, `\`) && i < len(tmpElements)-1 {
				elements = append(elements, "'"+strings.TrimSuffix(elem, `\`))
				quoted = true
			} else {
				elements = append(elements, elem)
			}
		}
	}

	sanitizedKey := strings.Join(elements, ".")
	if !strings.HasPrefix(sanitizedKey, "$.") {
		sanitizedKey = "$." + sanitizedKey
	}

	if sanitizedKey != key {
		logrus.Warningf("current yaml key is %q and should be updated to %q", key, sanitizedKey)
	}

	return sanitizedKey

}
