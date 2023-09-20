package engine

import (
	"github.com/updatecli/updatecli/pkg/core/registry"
)

// PullFromRegistry retrieves an Updateli policy from an OCI registry.
func (e *Engine) PullFromRegistry(policyReference string, disableTLS bool) error {

	PrintTitle("Registry")

	_, _, _, err := registry.Pull(policyReference, disableTLS)
	if err != nil {
		return err
	}

	return nil
}

// PushToRegistry pushes an Updateli policy to an OCI registry.
func (e *Engine) PushToRegistry(manifests, valuesFiles, secretsFiles, policyReference []string, disableTLS bool, filestore string) error {

	PrintTitle("Registry")

	manifests = sanitizeUpdatecliManifestFilePath(manifests)
	err := registry.Push(manifests, valuesFiles, secretsFiles, policyReference, disableTLS, filestore)
	if err != nil {
		return err
	}

	return nil
}
