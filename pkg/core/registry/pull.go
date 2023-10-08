package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"

	spec "github.com/opencontainers/image-spec/specs-go/v1"
)

// Pull pulls an OCI image from a registry.
func Pull(ociName string, disableTLS bool) (manifests []string, values []string, secrets []string, err error) {

	logrus.Infof("Pulling Updatecli policy %q\n", ociName)

	ref, err := registry.ParseReference(ociName)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("parse reference: %w", err)
	}

	if ref.Reference == ociLatestTag || ref.Reference == "" {
		ref.Reference, err = getLatestTagSortedBySemver(ref.Registry+"/"+ref.Repository, disableTLS)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("get latest tag sorted by semver: %w", err)
		}
		logrus.Debugf("Latest tag founded %s", ref.String())
	}

	// 1. Connect to a remote repository
	ctx := context.Background()

	repo, err := remote.NewRepository(ociName)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("new repository: %w", err)
	}

	if disableTLS {
		logrus.Debugln("TLS connection is disabled")
		repo.PlainHTTP = true
	}

	// 2. Get credentials from the docker credential store
	if err := getCredentialsFromDockerStore(repo); err != nil {
		return nil, nil, nil, fmt.Errorf("credstore from docker: %w", err)
	}

	// 2.5 Get remote manifest digest

	remoteManifestSpec, _, err := oras.Fetch(ctx, repo, ref.String(), oras.DefaultFetchOptions)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("fetch: %w", err)
	}

	// Create a file store
	store := filepath.Join(getReferencePath(remoteManifestSpec.Digest.String())...)
	fs, err := file.New(store)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("create file store: %w", err)
	}
	defer fs.Close()

	// 3. Copy from the remote repository to the file store
	manifestDescriptor, err := oras.Copy(ctx, repo, ref.Reference, fs, ref.Reference, oras.DefaultCopyOptions)
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

	logrus.Debugf("policy successfully pulled in %s", store)

	return manifests, values, secrets, nil
}

// getReferencePath returns the path to the file store for a given reference.
func getReferencePath(ref string) []string {
	refPath := []string{
		os.TempDir(),
		"updatecli",
		"store",
	}
	refPath = append(refPath, strings.TrimPrefix(ref, "sha256:"))

	return refPath
}
