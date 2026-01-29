package flux

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchFluxFiles(t *testing.T) {
	f, err := New(
		Spec{}, "testdata", "", "")

	require.NoError(t, err)

	err = f.searchFluxFiles("testdata", defaultFluxFiles[:])
	if err != nil {
		t.Error(err)
	}

	expectedHelmReleaseFile := []string{
		"testdata/helmrelease/multi-release/multi-helmrelease.yaml",
		"testdata/helmrelease/oci/helmrelease.yaml",
		"testdata/helmrelease/oci-combined/helmrelease-helmrepository.yaml",
		"testdata/helmrelease/simple/helmrelease.yaml",
		"testdata/helmrelease/simple-combined/helmrelease-helmrepository.yaml",
	}

	expectedOCIRepositoryFile := []string{
		"testdata/ociRepository/example.yaml",
		"testdata/ociRepository-latest/example.yaml",
	}

	assert.Equal(t, expectedHelmReleaseFile, f.helmReleaseFiles)
	assert.Equal(t, expectedOCIRepositoryFile, f.ociRepositoryFiles)
}
