package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	spec "github.com/opencontainers/image-spec/specs-go/v1"
	credentials "github.com/oras-project/oras-credentials-go"
	"github.com/sirupsen/logrus"

	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
)

// Pull pulls an OCI image from a registry.
func Pull(ociName string, disableTLS bool) (manifests []string, values []string, secrets []string, err error) {

	ref, err := registry.ParseReference(ociName)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("parse reference: %w", err)
	}

	if ref.Reference == "" {
		ref.Reference = ociDefaultTag
	}

	// Create a file store
	store := filepath.Join(getReferencePath(ref.String())...)
	fs, err := file.New(store)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create file store: %w", err)
	}
	defer fs.Close()

	// 1. Connect to a remote repository
	ctx := context.Background()

	repo, err := remote.NewRepository(ociName)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("new repository: %w", err)
	}

	if disableTLS {
		repo.PlainHTTP = true
	}

	// 2. Get credentials from the docker credential store
	storeOpts := credentials.StoreOptions{}
	credStore, err := credentials.NewStoreFromDocker(storeOpts)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("credstore from docker: %w", err)
	}

	repo.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.DefaultCache,
		Credential: credentials.Credential(credStore),
	}

	// 3. Copy from the remote repository to the file store
	tag := ociDefaultTag
	manifestDescriptor, err := oras.Copy(ctx, repo, tag, fs, tag, oras.DefaultCopyOptions)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("copy: %w", err)
	}

	manifestData, err := content.FetchAll(ctx, fs, manifestDescriptor)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("fetch manifest: %w", err)
	}

	spec := spec.Manifest{}
	err = json.Unmarshal(manifestData, &spec)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("unmarshal manifest: %w", err)
	}

	for _, layer := range spec.Layers {
		switch layer.MediaType {
		case updatecliManifestMediaType:
			if title, ok := layer.Annotations["org.opencontainers.image.title"]; ok && title != "" {
				manifests = append(manifests, filepath.Join(store, title))
			}

		case updatecliValueMediaType:
			if title, ok := layer.Annotations["org.opencontainers.image.title"]; ok && title != "" {
				values = append(values, filepath.Join(store, title))
			}

		case updatecliSecretMediaType:
			if title, ok := layer.Annotations["org.opencontainers.image.title"]; ok && title != "" {
				secrets = append(secrets, filepath.Join(store, title))
			}

		default:
			logrus.Warningf("unknown media type: %q\n", layer.MediaType)
		}
	}

	logrus.Debugf("Manifests:\n")
	for _, manifest := range manifests {
		logrus.Debugf("\t*%q\n", manifest)
	}

	if len(values) > 0 {
		logrus.Debugf("Values:\n")
		for _, value := range values {
			logrus.Debugf("\t*%q\n", value)
		}
	}

	if len(secrets) > 0 {
		logrus.Debugf("Secrets:\n")
		for _, secret := range secrets {
			logrus.Debugf("\t*%q\n", secret)
		}
	}

	return manifests, values, secrets, nil
}

// getReferencePath returns the path to the file store for a given reference.
func getReferencePath(ref string) []string {
	refArray := []string{
		os.TempDir(),
		"updatecli",
		"store",
	}
	refArray = append(
		refArray,
		strings.Split(
			strings.ReplaceAll(
				strings.ReplaceAll(
					ref, ":", "/"),
				".", "/"),
			"/")...)

	return refArray
}
