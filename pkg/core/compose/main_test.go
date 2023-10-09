package compose

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/engine"
)

func TestGetPolicies(t *testing.T) {

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
						filepath.Join("/", "tmp", "updatecli", "store", "7aaff2727eef42f7d0add2d5ed3fd83f74a125420682bec7e4bc8835bb28e833", "updatecli.d", "default.tpl"),
					},
					Values: []string{
						filepath.Join("/", "tmp", "updatecli", "store", "7aaff2727eef42f7d0add2d5ed3fd83f74a125420682bec7e4bc8835bb28e833", "values.yaml"),
					},
				},
			},
			expectedEnv: map[string]string{
				"UPDATECLI_TEST_GITHUB_TOKEN": "my super secret token",
				"UPDATECLI_TEST_GITHUB_ACTOR": "me",
			},
		},
	}

	for _, data := range testdata {
		t.Run(data.name, func(t *testing.T) {
			updateCompose, err := New(data.file)
			require.NoError(t, err)

			gotManifests, err := updateCompose.GetPolicies(false)
			require.NoError(t, err)

			assert.Equal(t, data.expectedManifests, gotManifests)

			for key, expectedValue := range data.expectedEnv {
				gotValue := os.Getenv(key)
				assert.Equal(t, expectedValue, gotValue)
			}
		})
	}
}
