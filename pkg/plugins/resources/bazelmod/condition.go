package bazelmod

import (
	"fmt"
	"path/filepath"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition checks if the version in MODULE.bazel matches the expected version
func (b *Bazelmod) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	// Use source as the expected version if provided
	expectedVersion := source
	if expectedVersion == "" {
		return false, "", fmt.Errorf("no version provided for condition check")
	}

	rootDir := ""
	if scm != nil {
		rootDir = scm.GetDirectory()
	}

	filePath := b.spec.File
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(rootDir, filePath)
	}

	// Check if file exists
	if !b.contentRetriever.FileExists(filePath) {
		return false, "", fmt.Errorf("MODULE.bazel file %q does not exist", filePath)
	}

	// Read file content
	content, err := b.contentRetriever.ReadAll(filePath)
	if err != nil {
		return false, "", fmt.Errorf("reading MODULE.bazel file: %w", err)
	}

	// Parse the MODULE.bazel file
	moduleFile, err := ParseModuleFile(content)
	if err != nil {
		return false, "", fmt.Errorf("parsing MODULE.bazel file: %w", err)
	}

	// Find the specified module
	dep := moduleFile.FindDepByName(b.spec.Module)
	if dep == nil {
		return false, "", fmt.Errorf("module %q not found in MODULE.bazel file %q", b.spec.Module, filePath)
	}

	// Compare versions
	if dep.Version == expectedVersion {
		return true, fmt.Sprintf("module %q version in %q is correctly set to %q",
			b.spec.Module,
			filePath,
			expectedVersion), nil
	}

	return false, fmt.Sprintf("module %q version in %q is incorrectly set to %q and should be %q",
		b.spec.Module,
		filePath,
		dep.Version,
		expectedVersion), nil
}
