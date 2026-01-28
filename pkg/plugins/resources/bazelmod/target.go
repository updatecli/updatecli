package bazelmod

import (
	"fmt"
	"path/filepath"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// Target updates the module version in MODULE.bazel file
func (b *Bazelmod) Target(source string, scm scm.ScmHandler, dryRun bool, resultTarget *result.Target) error {
	// Use source as the new version
	newVersion := source
	if newVersion == "" {
		return fmt.Errorf("no version provided for target update")
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

	// Check if version needs to be updated
	if dep.Version == newVersion {
		resultTarget.Information = dep.Version
		resultTarget.NewInformation = newVersion
		resultTarget.Result = result.SUCCESS
		resultTarget.Description = fmt.Sprintf("module %q version in %q is already set to %q",
			b.spec.Module,
			filePath,
			newVersion)
		return nil
	}

	// Update the version
	resultTarget.Information = dep.Version
	resultTarget.NewInformation = newVersion
	resultTarget.Result = result.ATTENTION
	resultTarget.Changed = true
	resultTarget.Description = fmt.Sprintf("module %q version in %q should be updated from %q to %q",
		b.spec.Module,
		filePath,
		dep.Version,
		newVersion)

	if dryRun {
		// Dry run: no changes to apply
		return nil
	}

	// Update the file content
	updatedContent, err := moduleFile.UpdateDepVersion(b.spec.Module, newVersion)
	if err != nil {
		return fmt.Errorf("updating module version: %w", err)
	}

	// Write the updated content back to the file
	err = b.contentRetriever.WriteToFile(updatedContent, filePath)
	if err != nil {
		return fmt.Errorf("writing updated MODULE.bazel file: %w", err)
	}

	resultTarget.Files = append(resultTarget.Files, filePath)
	resultTarget.Description = fmt.Sprintf("module %q version in %q updated from %q to %q",
		b.spec.Module,
		filePath,
		dep.Version,
		newVersion)

	return nil
}
