package terraform

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchTerraformLockFiles(t *testing.T) {
	dataset := []struct {
		name               string
		rootDir            string
		expectedFoundFiles []string
	}{
		{
			name:    "Default working scenario",
			rootDir: "testdata",
			expectedFoundFiles: []string{
				"testdata/.terraform.lock.hcl",
			},
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			foundFiles, err := searchTerraformLockFiles(d.rootDir)
			require.NoError(t, err)

			assert.Equal(t, foundFiles, d.expectedFoundFiles)
		})
	}
}

func TestGetTerraformLockContent(t *testing.T) {
	dataset := []struct {
		name              string
		file              string
		expectedProviders map[string]string
	}{
		{
			name: "Default working scenario",
			file: "testdata/.terraform.lock.hcl",
			expectedProviders: map[string]string{
				"registry.terraform.io/hashicorp/aws":       "5.9.0",
				"registry.terraform.io/hashicorp/cloudinit": "2.3.2",
			},
		},
	}

	for _, d := range dataset {
		t.Run(d.name, func(t *testing.T) {
			foundFiles, err := getTerraformLockContent(d.file)
			require.NoError(t, err)

			assert.Equal(t, foundFiles, d.expectedProviders)
		})
	}
}
