package engine

import (
	"github.com/updatecli/updatecli/pkg/core/registry"
)

// PullFromRegistry retrieves an Updatecli policy from an OCI registry.
func (e *Engine) PullFromRegistry(policyReference string, disableTLS bool) (err error) {

	PrintTitle("Registry")

	//nolint:dogsled
	_, _, _, err = registry.Pull(policyReference, disableTLS)
	if err != nil {
		return err
	}

	return nil
}

// PushToRegistry pushes an Updatecli policy to an OCI registry.
func (e *Engine) PushToRegistry(manifests, valuesFiles, secretsFiles, policyReference []string, disableTLS bool, policyMetadataFile, fileStore string) error {

	PrintTitle("Registry")

	manifests = sanitizeUpdatecliManifestFilePath(manifests)
	err := registry.Push(policyMetadataFile, manifests, valuesFiles, secretsFiles, policyReference, disableTLS, fileStore)
	if err != nil {
		return err
	}

	return nil
}
