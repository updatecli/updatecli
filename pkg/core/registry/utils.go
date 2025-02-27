package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/sirupsen/logrus"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/credentials"
	"oras.land/oras-go/v2/registry/remote/retry"
)

// getLatestTagSortedBySemver returns the latest tag sorted by semver
func getLatestTagSortedBySemver(refName string, disableTLS bool) (string, error) {

	repo, err := remote.NewRepository(refName)
	if err != nil {
		return "", fmt.Errorf("query repository: %w", err)
	}

	if disableTLS {
		repo.PlainHTTP = true
	}

	if err := getCredentialsFromDockerStore(repo); err != nil {
		return "", fmt.Errorf("credstore from docker: %w", err)
	}

	ctx := context.Background()
	ctx = auth.AppendRepositoryScope(ctx, repo.Reference, auth.ActionPull, auth.ActionPush)

	tags, err := registry.Tags(ctx, repo)

	if err != nil {
		return "", fmt.Errorf("get tags: %w", err)
	}

	result := []*semver.Version{}
	for i := range tags {
		s, err := semver.NewVersion(tags[i])
		if err != nil {
			logrus.Debugf("Ignoring tag %q - %q", tags[i], err)
			continue
		}

		result = append(result, s)
	}

	if len(result) == 0 {
		return "", fmt.Errorf("no valid semver tags found")
	}

	sort.Sort(semver.Collection(result))
	sort.Sort(sort.Reverse(semver.Collection(result)))

	latestTag := result[0].Original()
	logrus.Debugf("latest tag identified %q", latestTag)

	return latestTag, nil
}

// getCredentialsFromDockerStore get the credentials from the docker credential store
func getCredentialsFromDockerStore(repo *remote.Repository) error {

	storeOpts := credentials.StoreOptions{
		DetectDefaultNativeStore: true,
	}
	credStore, err := credentials.NewStoreFromDocker(storeOpts)
	if err != nil {
		return fmt.Errorf("credstore from docker: %w", err)
	}

	repo.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.NewSingleContextCache(),
		Credential: credentials.Credential(credStore),
	}

	return nil
}

// FetchManifest fetches the OCI manifest from the remote repository
func FetchManifest(ociName string, disableTLS bool) (v1.Descriptor, error) {

	ref, err := registry.ParseReference(ociName)
	if err != nil {
		return v1.Descriptor{}, fmt.Errorf("parse reference: %w", err)
	}

	if ref.Reference == ociLatestTag || ref.Reference == "" {
		ref.Reference, err = getLatestTagSortedBySemver(ref.Registry+"/"+ref.Repository, disableTLS)
		if err != nil {
			return v1.Descriptor{}, fmt.Errorf("get latest tag sorted by semver: %w", err)
		}
	}

	// 1. Connect to a remote repository
	ctx := context.Background()

	repo, err := remote.NewRepository(ociName)
	if err != nil {
		return v1.Descriptor{}, fmt.Errorf("new repository: %w", err)
	}

	ctx = auth.AppendRepositoryScope(ctx, repo.Reference, auth.ActionPull, auth.ActionPush)

	if disableTLS {
		logrus.Debugln("TLS connection is disabled")
		repo.PlainHTTP = true
	}

	// 2. Get credentials from the docker credential store
	if err := getCredentialsFromDockerStore(repo); err != nil {
		return v1.Descriptor{}, fmt.Errorf("credstore from docker: %w", err)
	}

	remoteManifestSpec, remoteManifestReader, err := oras.Fetch(ctx, repo, ref.String(), oras.DefaultFetchOptions)
	if err != nil {
		return v1.Descriptor{}, fmt.Errorf("fetch remote manifest: %w", err)
	}

	defer remoteManifestReader.Close()

	remoteManifestData, err := content.ReadAll(remoteManifestReader, remoteManifestSpec)
	if err != nil {
		return v1.Descriptor{}, fmt.Errorf("fetch remote content: %w", err)
	}

	manifestData := v1.Descriptor{}
	err = json.Unmarshal(remoteManifestData, &manifestData)
	if err != nil {
		return v1.Descriptor{}, fmt.Errorf("unmarshal: %w", err)
	}

	return manifestData, nil
}
