package bazelmod

import (
	"fmt"
	"path/filepath"

	"github.com/updatecli/updatecli/pkg/core/result"
)

// Source reads the version of the specified module from MODULE.bazel
func (b *Bazelmod) Source(workingDir string, resultSource *result.Source) error {
	filePath := b.spec.File
	if !filepath.IsAbs(filePath) {
		filePath = filepath.Join(workingDir, filePath)
	}

	// Check if file exists
	if !b.contentRetriever.FileExists(filePath) {
		return fmt.Errorf("MODULE.bazel file %q does not exist", filePath)
	}

	// Read file content
	content, err := b.contentRetriever.ReadAll(filePath)
	if err != nil {
		return fmt.Errorf("reading MODULE.bazel file: %w", err)
	}

	// Parse the MODULE.bazel file
	moduleFile, err := ParseModuleFile(content)
	if err != nil {
		return fmt.Errorf("parsing MODULE.bazel file: %w", err)
	}

	// Find the specified module
	dep := moduleFile.FindDepByName(b.spec.Module)
	if dep == nil {
		return fmt.Errorf("module %q not found in MODULE.bazel file %q", b.spec.Module, filePath)
	}

	resultSource.Information = dep.Version
	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("version %q found for module %q in file %q",
		dep.Version,
		b.spec.Module,
		filePath)

	return nil
}
