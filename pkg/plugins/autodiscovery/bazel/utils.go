package bazel

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

const (
	// ModuleBazelFile is the name of the Bazel module file
	ModuleBazelFile string = "MODULE.bazel"
)

// findModuleFilesFromFS recursively searches for MODULE.bazel files in the given filesystem.
// It skips hidden directories like .git, .bazel, etc.
// Returns relative paths from the filesystem root.
func findModuleFilesFromFS(fsys fs.FS) ([]string, error) {
	foundFiles := []string{}

	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			logrus.Debugf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		if d.IsDir() {
			base := filepath.Base(path)
			if len(base) > 0 && base[0] == '.' && base != "." && base != ".." {
				return fs.SkipDir
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
		return nil, fmt.Errorf("walking filesystem: %w", err)
	}

	logrus.Debugf("%d MODULE.bazel file(s) found", len(foundFiles))
	for _, foundFile := range foundFiles {
		logrus.Debugf("    * %q", foundFile)
	}

	return foundFiles, nil
}

// findModuleFiles recursively searches for MODULE.bazel files starting from rootDir.
// It skips hidden directories like .git, .bazel, etc.
func findModuleFiles(rootDir string) ([]string, error) {
	logrus.Debugf("Looking for MODULE.bazel files in %q", rootDir)

	fsys := os.DirFS(rootDir)
	relativeFiles, err := findModuleFilesFromFS(fsys)
	if err != nil {
		return nil, fmt.Errorf("walking directory %q: %w", rootDir, err)
	}

	// Convert relative paths to absolute paths
	foundFiles := make([]string, len(relativeFiles))
	for i, relFile := range relativeFiles {
		absFile := filepath.Join(rootDir, relFile)
		foundFiles[i] = absFile
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
