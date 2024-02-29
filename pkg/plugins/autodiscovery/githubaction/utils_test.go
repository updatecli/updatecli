package githubaction

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchWorkflowFiles(t *testing.T) {
	g, err := New(
		Spec{
			RootDir: "testdata",
		}, "", "")

	require.NoError(t, err)

	err = g.searchWorkflowFiles("testdata", defaultWorkflowFiles[:])
	if err != nil {
		t.Error(err)
	}

	expectedWorkflowFiles := []string{
		"testdata/updatecli/.github/workflows/updatecli.yaml",
	}

	assert.Equal(t, expectedWorkflowFiles, g.workflowFiles)
}
