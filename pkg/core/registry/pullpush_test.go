package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"

	"github.com/stretchr/testify/require"
)

// TestPushPull is a test for the Push and Pull functions
func TestPushPullPolicy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "ghcr.io/project-zot/zot-linux-amd64:latest",
		ExposedPorts: []string{"5000/tcp"},
		WaitingFor:   wait.ForHTTP("/"),
	}
	zotC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := zotC.Terminate(ctx); err != nil {
			panic(err)
		}
	}()

	port, err := zotC.MappedPort(ctx, "5000")
	require.NoError(t, err)

	testData := []struct {
		name                      string
		toPushPolicyFile          string
		expectedPullManifestFiles []string
		expectedPullValuesFiles   []string
		expectedPullSecretsFiles  []string
		expectedPullAssetsFiles   []string
		toPushFileStore           string
		pushData                  PushData
	}{
		{
			name:                      "Validate that we can push and pull a policy using the latest tag, even thought the tag is ignored",
			expectedPullManifestFiles: []string{"testdata/venom.yaml"},
			expectedPullValuesFiles:   []string{"testdata/values.yaml"},
			expectedPullSecretsFiles:  []string{"testdata/secrets.yaml"},
			pushData: PushData{
				PolicyReferenceNames: []string{fmt.Sprintf("localhost:%d/myrepo:latest", port.Int())},
				PolicyMetadataFile:   "testdata/Policy.yaml",
				DisableTLS:           true,
				ManifestsFiles:       []string{"testdata/venom.yaml"},
				ValuesFiles:          []string{"testdata/values.yaml"},
				SecretsFiles:         []string{"testdata/secrets.yaml"},
			},
			toPushFileStore: ".",
		},
		{
			name:                      "Validate that we can push and pull a policy without tag",
			expectedPullManifestFiles: []string{"testdata/venom.yaml"},
			expectedPullValuesFiles:   []string{"testdata/values.yaml"},
			expectedPullSecretsFiles:  []string{"testdata/secrets.yaml"},
			pushData: PushData{
				PolicyReferenceNames: []string{fmt.Sprintf("localhost:%d/myrepo", port.Int())},
				PolicyMetadataFile:   "testdata/Policy.yaml",
				DisableTLS:           true,
				ManifestsFiles:       []string{"testdata/venom.yaml"},
				ValuesFiles:          []string{"testdata/values.yaml"},
				SecretsFiles:         []string{"testdata/secrets.yaml"},
			},
			toPushFileStore: ".",
		},
		{
			name:                      "Validate that we can push and pull a policy without tag from a different file store",
			expectedPullManifestFiles: []string{"testdata/venom.yaml"},
			expectedPullValuesFiles:   []string{"testdata/values.yaml"},
			expectedPullSecretsFiles:  []string{"testdata/secrets.yaml"},
			toPushFileStore:           ".",
			pushData: PushData{
				PolicyReferenceNames: []string{fmt.Sprintf("localhost:%d/myrepo", port.Int())},
				PolicyMetadataFile:   "testdata/Policy.yaml",
				DisableTLS:           true,
				ManifestsFiles:       []string{"testdata/venom.yaml"},
				ValuesFiles:          []string{"testdata/values.yaml"},
				SecretsFiles:         []string{"testdata/secrets.yaml"},
			},
		},
		{
			name:                      "Validate that we can push and pull a policy with assets",
			expectedPullManifestFiles: []string{"testdata/venom.yaml"},
			expectedPullValuesFiles:   []string{"testdata/values.yaml"},
			expectedPullSecretsFiles:  []string{"testdata/secrets.yaml"},
			expectedPullAssetsFiles:   []string{"testdata/asset.sh"},
			toPushFileStore:           ".",
			pushData: PushData{
				PolicyReferenceNames: []string{fmt.Sprintf("localhost:%d/myrepo", port.Int())},
				DisableTLS:           true,
				PolicyMetadataFile:   "testdata/Policy.yaml",
				AssetsFiles:          []string{"testdata/asset.sh"},
				ManifestsFiles:       []string{"testdata/venom.yaml"},
				ValuesFiles:          []string{"testdata/values.yaml"},
				SecretsFiles:         []string{"testdata/secrets.yaml"},
			},
		},
	}

	for _, data := range testData {

		t.Run(data.name, func(t *testing.T) {
			err = Push(data.pushData)
			require.NoError(t, err)

			err = Push(data.pushData)
			require.NoError(t, err)

			gotManifests, gotValues, gotSecrets, err := Pull(
				data.pushData.PolicyReferenceNames[0],
				data.pushData.DisableTLS,
			)
			require.NoError(t, err)

			expectedManifest, expectedValues, expectedSecrets, err := sanitizeDirPath(
				data.pushData.PolicyReferenceNames[0],
				data.pushData.DisableTLS,
				data.expectedPullManifestFiles,
				data.expectedPullValuesFiles,
				data.expectedPullSecretsFiles,
			)
			require.NoError(t, err)

			require.Equal(t, expectedManifest, gotManifests)
			require.Equal(t, expectedValues, gotValues)
			require.Equal(t, expectedSecrets, gotSecrets)
		})
	}
}

func sanitizeDirPath(policyRef string, disableTLS bool, manifestFiles, valuesFiles, secretsFiles []string) ([]string, []string, []string, error) {

	ref, err := registry.ParseReference(policyRef)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("parse reference: %w", err)
	}

	if ref.Reference == ociLatestTag || ref.Reference == "" {
		ref.Reference, err = getLatestTagSortedBySemver(ref.Registry+"/"+ref.Repository, disableTLS)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("get latest tag sorted by semver: %w", err)
		}
	}

	// 1. Connect to a remote repository
	ctx := context.Background()

	repo, err := remote.NewRepository(policyRef)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("new repository: %w", err)
	}

	ctx = auth.AppendRepositoryScope(ctx, repo.Reference, auth.ActionPull, auth.ActionPush)

	if disableTLS {
		repo.PlainHTTP = true
	}

	// 2. Get credentials from the docker credential store
	if err := getCredentialsFromDockerStore(repo); err != nil {
		return nil, nil, nil, fmt.Errorf("ini repo settings: %w", err)
	}

	// 2.5 Get remote manifest digest

	remoteManifestSpec, _, err := oras.Fetch(ctx, repo, ref.String(), oras.DefaultFetchOptions)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("fetch: %w", err)
	}

	dirPath := []string{
		os.TempDir(),
		"updatecli",
		"store",
		strings.TrimPrefix(remoteManifestSpec.Digest.String(), "sha256:"),
	}

	addPrefix := func(files []string) {
		for i := range files {
			files[i] = filepath.Join(append(dirPath, files[i])...)
		}
	}

	addPrefix(manifestFiles)
	addPrefix(valuesFiles)
	addPrefix(secretsFiles)

	return manifestFiles, valuesFiles, secretsFiles, nil
}
