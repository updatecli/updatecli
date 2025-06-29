package engine

import (
	"path/filepath"

	"github.com/sirupsen/logrus"
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
func (e *Engine) PushToRegistry(manifests, valuesFiles, secretsFiles, policyReference []string, disableTLS bool, policyMetadataFile, fileStore string, overwrite bool) error {

	PrintTitle("Registry")

	joinWithFileStore := func(files []string) []string {
		for i, file := range files {
			if !filepath.IsAbs(file) {
				files[i] = filepath.Join(fileStore, file)
			}
		}
		return files
	}

	relativeFromFileStore := func(files []string) []string {
		for i, file := range files {
			relPath, err := filepath.Rel(fileStore, file)
			if err != nil {
				logrus.Errorf("Unable to get relative path from %s to %s", fileStore, file)
				continue
			}
			files[i] = relPath
		}
		return files
	}

	// If policyMetadataFile is not an absolute path then we assume it is relative to fileStore
	if !filepath.IsAbs(policyMetadataFile) {
		policyMetadataFile = filepath.Join(fileStore, policyMetadataFile)
	}

	joinWithFileStore(manifests)

	manifestFiles, partialFiles := sanitizeUpdatecliManifestFilePath(manifests)

	manifests = append(manifests, manifestFiles...)
	manifests = append(manifests, partialFiles...)

	relativeFromFileStore(manifests)

	err := registry.Push(policyMetadataFile, manifests, valuesFiles, secretsFiles, policyReference, disableTLS, fileStore, overwrite)
	if err != nil {
		return err
	}

	return nil
}
