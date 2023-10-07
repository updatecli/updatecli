package compose

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/engine"
	"github.com/updatecli/updatecli/pkg/core/tmp"
)

func TestGetPolicies(t *testing.T) {

	tmpDir := tmp.Directory

	testdata := []struct {
		name              string
		file              string
		expectedManifests []engine.Manifest
		expectedEnv       map[string]string
	}{
		{
			name: "Test getPolicies with environment variables",
			file: "testdata/update-compose.yaml",
			expectedManifests: []engine.Manifest{
				{
					Manifests: []string{
						filepath.Join(tmpDir, "store", "ghcr", "io", "olblak", "policies", "updatecli", "latest", "updatecli", "updatecli.d", "updatecli.yaml"),
					},
				},
			},
			expectedEnv: map[string]string{
				"UPDATECLI_TEST_GITHUB_TOKEN": "my super secret token",
				"UPDATECLI_TEST_GITHUB_ACTOR": "me",
			},
		},
	}

	for i := range testdata {
		updateCompose, err := New(testdata[i].file)
		require.NoError(t, err)

		gotManifests, err := updateCompose.GetPolicies(false)
		require.NoError(t, err)

		assert.Equal(t, testdata[i].expectedManifests, gotManifests)

		for key, expectedValue := range testdata[i].expectedEnv {
			gotValue := os.Getenv(key)
			assert.Equal(t, expectedValue, gotValue)
		}
	}
}
