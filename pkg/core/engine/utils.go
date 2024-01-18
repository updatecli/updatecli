package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

/*
sanitizeUpdatecliManifestFilePath receives a list of files (directory or file) and returns a list of files that could be accepted by Updatecli.
*/
func sanitizeUpdatecliManifestFilePath(rawFilePaths []string) (sanitizedFilePaths []string) {
	for _, r := range rawFilePaths {
		err := filepath.Walk(r, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				logrus.Errorf("\n%s File %s: %s\n", result.FAILURE, path, err)
				return fmt.Errorf("unable to walk %q: %s", path, err)
			}
			if info.Mode().IsRegular() {
				sanitizedFilePaths = append(sanitizedFilePaths, path)
			}
			return nil
		})

		if err != nil {
			logrus.Errorf("err - %s", err)
		}
	}

	// Remove duplicates manifest files
	result := []string{}
	exist := map[string]bool{}

	for v := range sanitizedFilePaths {
		if !exist[sanitizedFilePaths[v]] {
			exist[sanitizedFilePaths[v]] = true
			result = append(result, sanitizedFilePaths[v])
		}
	}

	return result
}

// PrintTitle print a title
func PrintTitle(title string) {
	logrus.Infof("\n\n%s\n", strings.Repeat("+", len(title)+4))
	logrus.Infof("+ %s +\n", strings.ToTitle(title))
	logrus.Infof("%s\n\n", strings.Repeat("+", len(title)+4))
}
