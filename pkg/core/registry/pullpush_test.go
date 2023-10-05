package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"context"

	credentials "github.com/oras-project/oras-credentials-go"
	"github.com/sirupsen/logrus"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"

	"github.com/stretchr/testify/require"
)

// TestPushPull is a test for the Push and Pull functions
func TestPushPullPolicy(t *testing.T) {
	//if testing.Short() {
	//	t.Skip("skipping integration test")
	//}

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

	err = Push(
		"testdata/Policy.yaml",
		[]string{
			"testdata/venom.yaml",
		},
		[]string{
			"testdata/values.yaml",
		},
		[]string{
			"testdata/secrets.yaml",
		},
		[]string{fmt.Sprintf("localhost:%s/myrepo:latest", port)},
		true,
		"")
	require.NoError(t, err)

	gotManifests, gotValues, gotSecrets, err := Pull(
		fmt.Sprintf("localhost:%s/myrepo:latest", port),
		true)
	require.NoError(t, err)

	expectedManifest := []string{
		"testdata/venom.yaml",
	}

	expectedValues := []string{
		"testdata/values.yaml",
	}

	expectedSecrets := []string{
		"testdata/secrets.yaml",
	}

	expectedManifest, expectedValues, expectedSecrets, err = sanitizeDirPath(fmt.Sprintf("localhost:%s/myrepo:latest", port), true, expectedManifest, expectedValues, expectedSecrets)
	require.NoError(t, err)

	require.Equal(t, expectedManifest, gotManifests)
	require.Equal(t, expectedValues, gotValues)
	require.Equal(t, expectedSecrets, gotSecrets)
}

func sanitizeDirPath(policyRef string, disableTLS bool, manifestFiles, valueFiles, secretfiles []string) ([]string, []string, []string, error) {
	// 1. Connect to a remote repository
	ctx := context.Background()

	repo, err := remote.NewRepository(policyRef)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("new repository: %w", err)
	}

	if disableTLS {
		logrus.Debugln("TLS connection is disabled")
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

	// 2.5 Get remote manifest digest

	remoteManifestSpec, _, err := oras.Fetch(ctx, repo, policyRef, oras.DefaultFetchOptions)
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
	addPrefix(valueFiles)
	addPrefix(secretfiles)

	return manifestFiles, valueFiles, secretfiles, nil
}
