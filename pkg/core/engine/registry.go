package engine

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/updatecli/updatecli/pkg/core/registry"
)

// PullFromRegistry retrieves an Updatecli policy from an OCI registry.
func (e *Engine) PullFromRegistry(policyReference string, disableTLS bool) (err error) {

	PrintTitle("Registry")

	_, err = registry.Pull(policyReference, disableTLS)
	if err != nil {
		return err
	}

	return nil
}

// PushToRegistry pushes an Updatecli policy to an OCI registry.
func (e *Engine) PushToRegistry(manifests, valuesFiles, secretsFiles, assetsFiles, policyReference []string, disableTLS bool, policyMetadataFile, fileStore string, overwrite bool) error {

	PrintTitle("Registry")

	// filepath.Rel needs both paths to be absolute, or both relative,
	// so fileStore must be absolute before comparing it with absolute input paths.
	absFileStore, err := filepath.Abs(fileStore)
	if err != nil {
		return fmt.Errorf("get absolute path of file store %q: %w", fileStore, err)
	}
	fileStore = absFileStore

	// Every file is stored in the policy under its path relative to fileStore,
	// so a manifest using "relativepaths: manifest" finds its assets at the
	// same place once pulled.
	relativeToFileStore := func(files []string) ([]string, error) {
		result := make([]string, 0, len(files))
		for _, file := range files {
			if !filepath.IsAbs(file) {
				file = filepath.Join(fileStore, file)
			}
			relPath, err := filepath.Rel(fileStore, file)
			if err != nil {
				return nil, fmt.Errorf("get path of %q relative to %q: %w", file, fileStore, err)
			}
			if relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
				return nil, fmt.Errorf("file %q is outside of %q", file, fileStore)
			}
			result = append(result, relPath)
		}
		return result, nil
	}

	// If policyMetadataFile is not an absolute path then we assume it is relative to fileStore
	if !filepath.IsAbs(policyMetadataFile) {
		policyMetadataFile = filepath.Join(fileStore, policyMetadataFile)
	}

	for i, file := range manifests {
		if !filepath.IsAbs(file) {
			manifests[i] = filepath.Join(fileStore, file)
		}
	}

	manifestFiles, partialFiles := sanitizeUpdatecliManifestFilePath(manifests)

	manifests = append(manifests, manifestFiles...)
	manifests = append(manifests, partialFiles...)

	if manifests, err = relativeToFileStore(manifests); err != nil {
		return fmt.Errorf("manifests: %w", err)
	}
	if valuesFiles, err = relativeToFileStore(valuesFiles); err != nil {
		return fmt.Errorf("values: %w", err)
	}
	if secretsFiles, err = relativeToFileStore(secretsFiles); err != nil {
		return fmt.Errorf("secrets: %w", err)
	}
	if assetsFiles, err = relativeToFileStore(assetsFiles); err != nil {
		return fmt.Errorf("assets: %w", err)
	}

	err = registry.Push(registry.PushData{
		PolicyMetadataFile:   policyMetadataFile,
		ManifestsFiles:       manifests,
		ValuesFiles:          valuesFiles,
		SecretsFiles:         secretsFiles,
		AssetsFiles:          assetsFiles,
		PolicyReferenceNames: policyReference,
		DisableTLS:           disableTLS,
		FileStore:            fileStore,
		Overwrite:            overwrite,
	})
	if err != nil {
		return err
	}

	return nil
}
