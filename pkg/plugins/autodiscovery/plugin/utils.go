package plugin

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

func getAllowedPaths(rootDir string, allowedPaths []string) map[string]string {
	result := map[string]string{}

	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}

	if rootDir == "" {
		rootDir = wd
	}

	getAbsolutePath := func(path string) string {
		if !filepath.IsAbs(path) {
			logrus.Debugf("Relative path %q detected, converting to absolute path based on working directory %q\n", path, rootDir)
			absPath := filepath.Join(rootDir, path)
			return absPath
		}
		return path
	}

	for _, path := range allowedPaths {

		split := strings.Split(path, ":")
		switch len(split) {
		case 1:
			result[getAbsolutePath(path)] = path
		case 2:
			result[getAbsolutePath(split[0])] = split[1]
		default:
			result[getAbsolutePath(split[0])] = strings.Join(split[1:], ":")
		}
	}
	return result
}
