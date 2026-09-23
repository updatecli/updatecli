package helm

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/utils/pathresolver"
)

// TestGetRepoIndexFromFileAbsoluteURLWithScm checks that, with an scm, an absolute file://
// repository is read under the checkout and never from the host.
func TestGetRepoIndexFromFileAbsoluteURLWithScm(t *testing.T) {
	checkout := t.TempDir()
	repoDir := filepath.Join(checkout, "charts")
	require.NoError(t, os.MkdirAll(repoDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repoDir, "index.yaml"), []byte("apiVersion: v1\nentries: {}\n"), 0o600))

	c, err := New(Spec{URL: "file:///charts", Name: "example"})
	require.NoError(t, err)

	index, err := c.GetRepoIndexFromFile(pathresolver.New(&scm.MockScm{WorkingDir: checkout}, ""))
	require.NoError(t, err)
	assert.Equal(t, "v1", index.APIVersion)
}
