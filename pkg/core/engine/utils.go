package engine

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/result"
)

/*
sanitizeUpdatecliManifestFilePath receives a list of files (directory or file) and returns a list of files that could be accepted by Updatecli.
*/
func sanitizeUpdatecliManifestFilePath(rawFilePaths []string, rootDirPath string) (sanitizedFilePaths []string) {
	if len(rawFilePaths) == 0 {
		// Updatecli tries to load the file updatecli.yaml if no manifest provided
		// If updatecli.yaml doesn't exists then Updatecli parses the directory updatecli.d for any manifests.
		// if there is no manifests in the directory updatecli.d then Updatecli returns no manifest files.
		_, err := os.Stat(config.DefaultConfigFilename)
		if !errors.Is(err, os.ErrNotExist) {
			logrus.Debugf("Default Updatecli manifest detected %q", config.DefaultConfigFilename)
			return []string{config.DefaultConfigFilename}
		}

		fs, err := os.Stat(config.DefaultConfigDirname)
		if errors.Is(err, os.ErrNotExist) {
			return []string{}
		}

		if fs.IsDir() {
			logrus.Debugf("Default Updatecli manifest directory detected %q", config.DefaultConfigDirname)
			rawFilePaths = []string{config.DefaultConfigDirname}
		}
	}

	for _, r := range rawFilePaths {
		r = filepath.Join(rootDirPath, r)
		err := filepath.Walk(r, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				logrus.Errorf("\n%s File %s: %s\n", result.FAILURE, path, err)
				return fmt.Errorf("unable to walk %q: %s", path, err)
			}
			if info.Mode().IsRegular() {
				if rootDirPath != "" {
					tmpPath := path
					path, err = filepath.Rel(rootDirPath, path)
					if err != nil {
						return fmt.Errorf("unable to get relative path for %q: %s", tmpPath, err)
					}
				}
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
