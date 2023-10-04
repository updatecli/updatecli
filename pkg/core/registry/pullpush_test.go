package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

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

	dirPath := []string{
		os.TempDir(),
		"updatecli",
		"store",
		"cf3da6a8f1e1073faef196e9bf168e43c2e25cb6291e1bfe1ec3b1c4918a7e5",
	}

	expectedManifest := []string{
		filepath.Join(append(dirPath, "tesdata/venom.yaml")...),
	}

	expectedValues := []string{
		filepath.Join(append(dirPath, "tesdata/values.yaml")...),
	}

	expectedSecrets := []string{
		filepath.Join(append(dirPath, "tesdata/secrets.yaml")...),
	}

	require.Equal(t, expectedManifest, gotManifests)
	require.Equal(t, expectedValues, gotValues)
	require.Equal(t, expectedSecrets, gotSecrets)
}
