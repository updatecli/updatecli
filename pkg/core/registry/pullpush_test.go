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
		toPushPolicyName          []string
		toPushPolicyFile          string
		toPushManifestFiles       []string
		toPushValueFiles          []string
		toPushSecretFiles         []string
		toPushFileStore           string
		expectedPullManifestFiles []string
		expectedPullValuesFiles   []string
		expectedPullSecretsFiles  []string
		disableTLS                bool
		overwrite                 bool
	}{
		{
			name:                      "Validate that we can push and pull a policy using the latest tag, even thought the tag is ignored",
			toPushPolicyName:          []string{fmt.Sprintf("localhost:%d/myrepo:latest", port.Int())},
			disableTLS:                true,
			toPushPolicyFile:          "testdata/Policy.yaml",
			toPushManifestFiles:       []string{"testdata/venom.yaml"},
			toPushValueFiles:          []string{"testdata/values.yaml"},
			toPushSecretFiles:         []string{"testdata/secrets.yaml"},
			expectedPullManifestFiles: []string{"testdata/venom.yaml"},
			expectedPullValuesFiles:   []string{"testdata/values.yaml"},
			expectedPullSecretsFiles:  []string{"testdata/secrets.yaml"},
		},
		{
			name:                      "Validate that we can push and pull a policy without tag",
			toPushPolicyName:          []string{fmt.Sprintf("localhost:%d/myrepo", port.Int())},
			disableTLS:                true,
			toPushPolicyFile:          "testdata/Policy.yaml",
			toPushManifestFiles:       []string{"testdata/venom.yaml"},
			toPushValueFiles:          []string{"testdata/values.yaml"},
			toPushSecretFiles:         []string{"testdata/secrets.yaml"},
			expectedPullManifestFiles: []string{"testdata/venom.yaml"},
			expectedPullValuesFiles:   []string{"testdata/values.yaml"},
			expectedPullSecretsFiles:  []string{"testdata/secrets.yaml"},
		},
		{
			name:                      "Validate that we can push and pull a policy without tag from a different file store",
			toPushPolicyName:          []string{fmt.Sprintf("localhost:%d/myrepo", port.Int())},
			disableTLS:                true,
			toPushPolicyFile:          "testdata/Policy.yaml",
			toPushManifestFiles:       []string{"testdata/venom.yaml"},
			toPushValueFiles:          []string{"testdata/values.yaml"},
			toPushSecretFiles:         []string{"testdata/secrets.yaml"},
			expectedPullManifestFiles: []string{"testdata/venom.yaml"},
			expectedPullValuesFiles:   []string{"testdata/values.yaml"},
			expectedPullSecretsFiles:  []string{"testdata/secrets.yaml"},
			toPushFileStore:           ".",
		},
	}

	for _, data := range testData {

		t.Run(data.name, func(t *testing.T) {
			err = Push(
				data.toPushPolicyFile,
				data.toPushManifestFiles,
				data.toPushValueFiles,
				data.toPushSecretFiles,
				data.toPushPolicyName,
				data.disableTLS,
				data.toPushFileStore,
				data.overwrite)
			require.NoError(t, err)

			err = Push(
				data.toPushPolicyFile,
				data.toPushManifestFiles,
				data.toPushValueFiles,
				data.toPushSecretFiles,
				data.toPushPolicyName,
				data.disableTLS,
				data.toPushFileStore,
				data.overwrite)
			require.NoError(t, err)

			gotManifests, gotValues, gotSecrets, err := Pull(
				data.toPushPolicyName[0],
				data.disableTLS,
			)
			require.NoError(t, err)

			expectedManifest, expectedValues, expectedSecrets, err := sanitizeDirPath(
				data.toPushPolicyName[0],
				data.disableTLS,
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
