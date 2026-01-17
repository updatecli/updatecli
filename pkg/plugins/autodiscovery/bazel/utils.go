package bazel

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

const (
	// ModuleBazelFile is the name of the Bazel module file
	ModuleBazelFile string = "MODULE.bazel"
)

// findModuleFiles recursively searches for MODULE.bazel files starting from rootDir.
// It skips hidden directories like .git, .bazel, etc.
func findModuleFiles(rootDir string) ([]string, error) {
	foundFiles := []string{}

	logrus.Debugf("Looking for MODULE.bazel files in %q", rootDir)

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logrus.Debugf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if d.IsDir() {
			base := filepath.Base(path)
			if len(base) > 0 && base[0] == '.' && base != "." && base != ".." {
				return filepath.SkipDir
			}
		}

		// Check if this is a MODULE.bazel file
		if !d.IsDir() && d.Name() == ModuleBazelFile {
			foundFiles = append(foundFiles, path)
			logrus.Debugf("Found MODULE.bazel file: %q", path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking directory %q: %w", rootDir, err)
	}

	logrus.Debugf("%d MODULE.bazel file(s) found", len(foundFiles))
	for _, foundFile := range foundFiles {
		relPath, err := filepath.Rel(rootDir, foundFile)
		if err == nil {
			logrus.Debugf("    * %q", relPath)
		} else {
			logrus.Debugf("    * %q", foundFile)
		}
	}

	return foundFiles, nil
}

// shouldIgnore checks if a module should be ignored based on matching rules
func shouldIgnore(moduleName, moduleVersion, filePath, rootDir string, rules MatchingRules) bool {
	if len(rules) == 0 {
		return false
	}

	relativePath, err := filepath.Rel(rootDir, filePath)
	if err != nil {
		relativePath = filePath
	}

	return rules.isMatchingRules(rootDir, relativePath, moduleName, moduleVersion)
}

// shouldInclude checks if a module should be included based on matching rules
func shouldInclude(moduleName, moduleVersion, filePath, rootDir string, rules MatchingRules) bool {
	if len(rules) == 0 {
		return true // If no Only rules, include everything
	}

	relativePath, err := filepath.Rel(rootDir, filePath)
	if err != nil {
		relativePath = filePath
	}

	return rules.isMatchingRules(rootDir, relativePath, moduleName, moduleVersion)
}
