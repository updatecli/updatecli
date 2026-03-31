package compose

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/engine/manifest"
)

func TestGetPolicies(t *testing.T) {

	testdata := []struct {
		name              string
		file              string
		expectedManifests [][]manifest.Manifest
		expectedEnv       map[string]string
		ignorePolicyIDs   []string
		onlyPolicyIDs     []string
	}{
		{
			name: "Test getPolicies with only rules",
			file: "testdata/policies/updatecli-compose.yaml",
			onlyPolicyIDs: []string{
				"scm_enabled",
			},
			expectedManifests: [][]manifest.Manifest{
				{
					{
						Manifests: []string{
							filepath.Join(os.TempDir(), "updatecli", "store", "7aaff2727eef42f7d0add2d5ed3fd83f74a125420682bec7e4bc8835bb28e833", "updatecli.d", "default.tpl"),
						},
						Values: []string{
							filepath.Join(os.TempDir(), "updatecli", "store", "7aaff2727eef42f7d0add2d5ed3fd83f74a125420682bec7e4bc8835bb28e833", "values.yaml"),
						},
						ValuesInline: []string{
							"scm:\n    enabled: true\n",
						},
					},
				},
			},
		},
		{
			name: "Test getPolicies with ignore rules",
			file: "testdata/policies/updatecli-compose.yaml",
			ignorePolicyIDs: []string{
				"scm_enabled",
			},
			expectedManifests: [][]manifest.Manifest{
				{
					{
						Manifests: []string{
							filepath.Join(os.TempDir(), "updatecli", "store", "7aaff2727eef42f7d0add2d5ed3fd83f74a125420682bec7e4bc8835bb28e833", "updatecli.d", "default.tpl"),
						},
						Values: []string{
							filepath.Join(os.TempDir(), "updatecli", "store", "7aaff2727eef42f7d0add2d5ed3fd83f74a125420682bec7e4bc8835bb28e833", "values.yaml"),
						},
						ValuesInline: []string{
							"scm:\n    enabled: false\n",
						},
					},
				},
			},
		},
		{
			name: "Test getPolicies with environment variables",
			file: "testdata/simple/updatecli-compose.yaml",
			expectedManifests: [][]manifest.Manifest{
				{
					{
						Manifests: []string{
							filepath.Join(os.TempDir(), "updatecli", "store", "7aaff2727eef42f7d0add2d5ed3fd83f74a125420682bec7e4bc8835bb28e833", "updatecli.d", "default.tpl"),
						},
						Values: []string{
							filepath.Join(os.TempDir(), "updatecli", "store", "7aaff2727eef42f7d0add2d5ed3fd83f74a125420682bec7e4bc8835bb28e833", "values.yaml"),
						},
					},
				},
			},
			expectedEnv: map[string]string{
				"UPDATECLI_TEST_GITHUB_TOKEN": "my super secret token",
				"UPDATECLI_TEST_GITHUB_ACTOR": "me",
			},
		},
		{
			name: "Test getPolicies from included compose file with environment variables",
			file: "testdata/multiple/updatecli-compose.yaml",
			expectedManifests: [][]manifest.Manifest{
				// The root compose file doesn't contain any policy
				{},
				{
					{
						Manifests: []string{
							filepath.Join(os.TempDir(), "updatecli", "store", "7aaff2727eef42f7d0add2d5ed3fd83f74a125420682bec7e4bc8835bb28e833", "updatecli.d", "default.tpl"),
						},
						Values: []string{
							filepath.Join(os.TempDir(), "updatecli", "store", "7aaff2727eef42f7d0add2d5ed3fd83f74a125420682bec7e4bc8835bb28e833", "values.yaml"),
						},
						ValuesInline: []string{
							"scm:\n    enabled: false\n",
						},
					},
				},
			},
			expectedEnv: map[string]string{
				"UPDATECLI_TEST_GITHUB_TOKEN": "my super secret token",
				"UPDATECLI_TEST_GITHUB_ACTOR": "me",
			},
		},
		{
			name: "Test getPolicies from included compose file with a circular reference",
			file: "testdata/circular/updatecli-compose.yaml",
			expectedManifests: [][]manifest.Manifest{
				// The root compose file doesn't contain any policy
				{},
				{
					{
						Manifests: []string{
							filepath.Join(os.TempDir(), "updatecli", "store", "7aaff2727eef42f7d0add2d5ed3fd83f74a125420682bec7e4bc8835bb28e833", "updatecli.d", "default.tpl"),
						},
						Values: []string{
							filepath.Join(os.TempDir(), "updatecli", "store", "7aaff2727eef42f7d0add2d5ed3fd83f74a125420682bec7e4bc8835bb28e833", "values.yaml"),
						},
						ValuesInline: []string{
							"scm:\n    enabled: false\n",
						},
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
			updatecliComposes, err := New(data.file, map[string]bool{})
			require.NoError(t, err)

			require.Equal(t, len(data.expectedManifests), len(updatecliComposes))

			for i := range updatecliComposes {
				c := updatecliComposes[i]
				gotManifests, err := c.GetPolicies(false, data.onlyPolicyIDs, data.ignorePolicyIDs)
				require.NoError(t, err)

				assert.Equal(t, data.expectedManifests[i], gotManifests)

				for key, expectedValue := range data.expectedEnv {
					gotValue := os.Getenv(key)
					assert.Equal(t, expectedValue, gotValue)
				}
			}

		})
	}
}
