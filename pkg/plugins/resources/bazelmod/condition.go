package bazelmod

import (
	"fmt"
	"path/filepath"

	"github.com/Masterminds/semver/v3"
	"github.com/sirupsen/logrus"
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

	// Compare versions using semantic version comparison if possible,
	// fall back to string comparison for non-semver versions
	versionsMatch := false
	depVer, depErr := semver.NewVersion(dep.Version)
	expectedVer, expectedErr := semver.NewVersion(expectedVersion)

	if depErr == nil && expectedErr == nil {
		// Both versions are valid semver, use semantic comparison
		versionsMatch = depVer.Equal(expectedVer)
		if versionsMatch {
			return true, fmt.Sprintf("module %q version in %q is correctly set to %q",
				b.spec.Module,
				filePath,
				expectedVersion), nil
		}
		logrus.Debugf("Semantic version comparison: %q != %q", dep.Version, expectedVersion)
	} else {
		// At least one version is not valid semver, use string comparison
		if depErr != nil {
			logrus.Debugf("Version %q is not a valid semantic version, using string comparison: %v", dep.Version, depErr)
		}
		if expectedErr != nil {
			logrus.Debugf("Version %q is not a valid semantic version, using string comparison: %v", expectedVersion, expectedErr)
		}
		versionsMatch = dep.Version == expectedVersion
		if versionsMatch {
			return true, fmt.Sprintf("module %q version in %q is correctly set to %q",
				b.spec.Module,
				filePath,
				expectedVersion), nil
		}
	}

	return false, fmt.Sprintf("module %q version in %q is incorrectly set to %q and should be %q",
		b.spec.Module,
		filePath,
		dep.Version,
		expectedVersion), nil
}
