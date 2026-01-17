package bazelregistry

import (
	"fmt"

	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

// Condition checks if a specific version exists in the Bazel Central Registry
func (b *Bazelregistry) Condition(source string, scm scm.ScmHandler) (pass bool, message string, err error) {
	if scm != nil {
		// SCM is not used for registry queries, but we don't error out
		// as it's a valid use case to check version existence
	}

	versionToCheck := source
	if len(versionToCheck) == 0 {
		return false, "", fmt.Errorf("no version provided for condition check")
	}

	// Fetch metadata from registry
	metadata, err := b.fetchModuleMetadata(b.spec.Module)
	if err != nil {
		return false, "", fmt.Errorf("fetching module metadata: %w", err)
	}

	// Check if version exists in the versions list
	for _, v := range metadata.Versions {
		if v == versionToCheck {
			// Check if version is yanked
			if reason, yanked := metadata.YankedVersions[versionToCheck]; yanked {
				return false, fmt.Sprintf("version %q exists but is yanked: %s", versionToCheck, reason), nil
			}
			return true, fmt.Sprintf("version %q is available in registry for module %q", versionToCheck, b.spec.Module), nil
		}
	}

	return false, fmt.Sprintf("version %q not found in registry for module %q", versionToCheck, b.spec.Module), nil
}
