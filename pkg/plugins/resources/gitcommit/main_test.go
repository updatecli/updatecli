package gitcommit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	resource, err := New(map[string]interface{}{
		"path":   "/tmp/repository",
		"branch": "main",
	})
	require.NoError(t, err)
	assert.Equal(t, "/tmp/repository", resource.spec.Path)
	assert.Equal(t, "main", resource.spec.Branch)
}

func TestReportConfig(t *testing.T) {
	depth := 1
	resource := &GitCommit{spec: Spec{
		Path:     "/tmp/repository",
		Branch:   "main",
		Depth:    &depth,
		URL:      "https://user:secret@example.com/owner/repository.git",
		Username: "user",
		Password: "secret",
	}}

	assert.Equal(t, Spec{
		Path:   "/tmp/repository",
		Branch: "main",
		Depth:  &depth,
		URL:    "https://****:****@example.com/owner/repository.git",
	}, resource.ReportConfig())
}
