package registry

import (
	"context"
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
	credentials "github.com/oras-project/oras-credentials-go"
	"github.com/sirupsen/logrus"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
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

	tags, err := registry.Tags(ctx, repo)
	if err != nil {
		return "", fmt.Errorf("get tags: %w", err)
	}

	result := make([]*semver.Version, len(tags))
	for i := range tags {
		s, err := semver.NewVersion(tags[i])
		if err != nil {
			logrus.Warningf("Can't parse tag %q as a semver version - %q", tags[i], err)
			continue
		}

		result[i] = s
	}

	if len(result) == 0 {
		return "", fmt.Errorf("no valid semver tags found")
	}
	sort.Sort(semver.Collection(result))

	return result[0].Original(), nil
}

// getCredentialsFromDockerStore get the credentials from the docker credential store
func getCredentialsFromDockerStore(repo *remote.Repository) error {

	storeOpts := credentials.StoreOptions{}
	credStore, err := credentials.NewStoreFromDocker(storeOpts)
	if err != nil {
		return fmt.Errorf("credstore from docker: %w", err)
	}

	repo.Client = &auth.Client{
		Client:     retry.DefaultClient,
		Cache:      auth.DefaultCache,
		Credential: credentials.Credential(credStore),
	}

	return nil
}
