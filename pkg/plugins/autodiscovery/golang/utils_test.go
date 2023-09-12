package golang

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchGoModFiles(t *testing.T) {

	dataset := []struct {
		name               string
		rootDir            string
		expectedFoundFiles []string
	}{
		{
			name:    "Default working scenario",
			rootDir: "test/testdata",
			expectedFoundFiles: []string{
				"test/testdata/noModule/go.mod",
			},
		},
	}

	pwd, err := os.Getwd()
	require.NoError(t, err)

	for _, d := range dataset {

		for i := range d.expectedFoundFiles {
			d.expectedFoundFiles[i] = path.Join(pwd, d.expectedFoundFiles[i])
		}

		t.Run(d.name, func(t *testing.T) {
			foundFiles, err := searchGoModFiles(d.rootDir)
			require.NoError(t, err)

			assert.Equal(t, d.expectedFoundFiles, foundFiles)
		})
	}
}
