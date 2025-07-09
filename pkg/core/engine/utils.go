package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// sanitizeUpdatecliManifestFilePath receives a list of files (directory or file)
// and returns both a list of files that could be accepted by Updatecli.
// a list of files that can be used as helpers.
func sanitizeUpdatecliManifestFilePath(rawFilePaths []string) (sanitizedFilePaths, sanitizedPartialPaths []string) {
	for _, rawFilePath := range rawFilePaths {
		rawFileInfo, err := os.Stat(rawFilePath)
		if err != nil {
			logrus.Error(fmt.Sprintf("Loading Updatecli manifest %q: %s", rawFilePath, err))
			continue
		}

		// If the manifest if a directory, then we walk trough it to find all manifest files
		// and partial files.
		if rawFileInfo.IsDir() {
			err = filepath.Walk(rawFilePath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					logrus.Errorf("\n%s File %s: %s\n", result.FAILURE, path, err)
					return fmt.Errorf("unable to walk %q: %s", path, err)
				}
				if info.Mode().IsRegular() {
					baseFile := filepath.Base(path)

					if strings.HasPrefix(baseFile, "_") {
						sanitizedPartialPaths = append(sanitizedPartialPaths, path)
					} else {
						sanitizedFilePaths = append(sanitizedFilePaths, path)
					}
				}
				return nil
			})
		}

		// If the manifest is a file, then we check any additional partial files
		// in the same directory that start with an underscore.
		if rawFileInfo.Mode().IsRegular() {
			manifestDirname := filepath.Dir(rawFilePath)
			dirEntries, err := os.ReadDir(manifestDirname) // Ensure the directory exists
			if err != nil {
				logrus.Errorf("unable to read directory %q: %s", manifestDirname, err)
				return nil, nil
			}

			for _, entry := range dirEntries {
				if entry.IsDir() {
					continue // Skip directories
				}

				baseFile := entry.Name()
				if strings.HasPrefix(baseFile, "_") {
					// If the file starts with an underscore, we consider it a partial file
					partialFilePath := filepath.Join(manifestDirname, baseFile)
					sanitizedPartialPaths = append(sanitizedPartialPaths, partialFilePath)
				}
			}

			sanitizedFilePaths = append(sanitizedFilePaths, rawFilePath)
		}

		if err != nil {
			logrus.Errorf("err - %s", err)
		}
	}

	trimDuplicate := func(input []string) []string {
		result := []string{}
		exist := map[string]bool{}

		for v := range input {
			if !exist[input[v]] {
				exist[input[v]] = true
				result = append(result, input[v])
			}
		}
		return result
	}

	return trimDuplicate(sanitizedFilePaths), trimDuplicate(sanitizedPartialPaths)
}

// PrintTitle print a title
func PrintTitle(title string) {
	logrus.Infof("\n\n%s\n", strings.Repeat("+", len(title)+4))
	logrus.Infof("+ %s +\n", strings.ToTitle(title))
	logrus.Infof("%s\n\n", strings.Repeat("+", len(title)+4))
}
