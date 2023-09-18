package registry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestPull is a test for the pull function
func TestPull(t *testing.T) {
	gotManifests, gotValues, gotSecrets, err := Pull(
		"localhost:5000/myrepo:latest",
		true)
	require.NoError(t, err)

	expectedManifest := []string{
		"/tmp/updatecli/store/localhost/5000/myrepo/latest/testdata/venom.yaml",
	}

	expectedValues := []string{
		"/tmp/updatecli/store/localhost/5000/myrepo/latest/testdata/values.yaml",
	}

	expectedSecrets := []string{
		"/tmp/updatecli/store/localhost/5000/myrepo/latest/testdata/secrets.yaml",
	}

	require.Equal(t, expectedManifest, gotManifests)
	require.Equal(t, expectedValues, gotValues)
	require.Equal(t, expectedSecrets, gotSecrets)
}
