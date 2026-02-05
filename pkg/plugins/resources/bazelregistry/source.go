package bazelregistry

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/result"
	"github.com/updatecli/updatecli/pkg/plugins/utils/version"
)

// Source queries the Bazel Central Registry and returns the appropriate version
func (b *Bazelregistry) Source(workingDir string, resultSource *result.Source) error {
	resultSource.Result = result.FAILURE

	// Fetch metadata from registry
	metadata, err := b.fetchModuleMetadata(b.spec.Module)
	if err != nil {
		return fmt.Errorf("fetching module metadata: %w", err)
	}

	if len(metadata.Versions) == 0 {
		return fmt.Errorf("no versions found for module %q", b.spec.Module)
	}

	// Filter out yanked versions
	availableVersions := make([]string, 0, len(metadata.Versions))
	yankedSet := make(map[string]bool)
	for version := range metadata.YankedVersions {
		yankedSet[version] = true
	}

	for _, v := range metadata.Versions {
		if !yankedSet[v] {
			availableVersions = append(availableVersions, v)
		}
	}

	if len(availableVersions) == 0 {
		return fmt.Errorf("all versions for module %q are yanked", b.spec.Module)
	}

	// Apply version filter
	var selectedVersion string
	// Check if this is the default "latest" filter (set by Init() when no filter specified)
	if b.versionFilter.Kind == version.LATESTVERSIONKIND && b.versionFilter.Pattern == version.LATESTVERSIONKIND {
		// Default "latest" - use semver to get latest with proper semantic sorting
		semverFilter := version.Filter{
			Kind:    version.SEMVERVERSIONKIND,
			Pattern: "*", // Any version = latest
		}
		semverFilter, err = semverFilter.Init()
		if err != nil {
			return fmt.Errorf("initializing semver filter: %w", err)
		}
		foundVersion, err := semverFilter.Search(availableVersions)
		if err != nil {
			// If semver fails (e.g., no valid semver versions), fall back to lexicographic latest
			// This handles edge cases where versions might not be semantic versions
			selectedVersion = availableVersions[len(availableVersions)-1]
		} else {
			selectedVersion = foundVersion.GetVersion()
		}
	} else {
		// Apply user-specified version filter
		foundVersion, err := b.versionFilter.Search(availableVersions)
		if err != nil {
			return fmt.Errorf("searching for version: %w", err)
		}
		selectedVersion = foundVersion.GetVersion()
	}

	if selectedVersion == "" {
		return fmt.Errorf("no version found matching filter for module %q", b.spec.Module)
	}

	resultSource.Information = selectedVersion
	resultSource.Result = result.SUCCESS
	resultSource.Description = fmt.Sprintf("version %q found for module %q", selectedVersion, b.spec.Module)

	return nil
}
