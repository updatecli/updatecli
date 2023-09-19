package registry

import (
	"fmt"
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

	expectedManifest := []string{
		fmt.Sprintf("/tmp/updatecli/store/localhost/%s/myrepo/latest/testdata/venom.yaml", port),
	}

	expectedValues := []string{
		fmt.Sprintf("/tmp/updatecli/store/localhost/%s/myrepo/latest/testdata/values.yaml", port),
	}

	expectedSecrets := []string{
		fmt.Sprintf("/tmp/updatecli/store/localhost/%s/myrepo/latest/testdata/secrets.yaml", port),
	}

	require.Equal(t, expectedManifest, gotManifests)
	require.Equal(t, expectedValues, gotValues)
	require.Equal(t, expectedSecrets, gotSecrets)
}
