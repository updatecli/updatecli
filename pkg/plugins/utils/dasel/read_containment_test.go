package dasel

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/text"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

// TestFileContent_ReadPathContainment is the regression test for
// GHSA-hj4x-hm4v-7wpw at the dasel sink shared by the json, toml and csv
// resources: a path escaping the working directory must be rejected before any
// file is accessed. ContentRetriever is intentionally left nil so the test
// fails loudly if containment ever stops short-circuiting the file access.
func TestFileContent_ReadPathContainment(t *testing.T) {
	workingDir := t.TempDir()

	tests := []struct {
		name     string
		filePath string
	}{
		{name: "absolute path", filePath: filepath.Join(t.TempDir(), "evil.toml")},
		{name: "dot dot traversal", filePath: filepath.Join("..", "..", "evil.toml")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := FileContent{
				DataType: "toml",
				FilePath: tt.filePath,
			}
			assert.Error(t, f.Read(utils.Resolver{BaseDir: workingDir, Boundary: workingDir}))
		})
	}
}

// TestFileContent_ReadLocalRunAcceptsAbsolutePath pins the counterpart of the containment
// rule: without an SCM checkout there is no boundary, so an absolute spec.file is a
// legitimate local path and must be read rather than rejected.
//
// It regressed when the containment fix started treating the process working directory as
// a boundary for the json, toml, csv and toolversions resources.
func TestFileContent_ReadLocalRunAcceptsAbsolutePath(t *testing.T) {
	absoluteFilePath := filepath.Join(t.TempDir(), "data.json")
	require.NoError(t, os.WriteFile(absoluteFilePath, []byte(`{"version":"1.0.0"}`), 0o600))

	f := FileContent{
		DataType:         "json",
		FilePath:         absoluteFilePath,
		ContentRetriever: &text.Text{},
	}

	require.NoError(t, f.Read(utils.Resolver{}))
	assert.Equal(t, absoluteFilePath, f.FilePath)
}

// TestFileContent_ReadIsIdempotent guards the base directory from being joined twice: Read
// mutates FilePath in place, so a second call must resolve from the manifest path again
// rather than from the already resolved one.
func TestFileContent_ReadIsIdempotent(t *testing.T) {
	baseDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(baseDir, "data.json"), []byte(`{"version":"1.0.0"}`), 0o600))

	f := FileContent{
		DataType:         "json",
		FilePath:         "data.json",
		ContentRetriever: &text.Text{},
	}

	resolver := utils.Resolver{BaseDir: baseDir}

	require.NoError(t, f.Read(resolver))
	firstFilePath := f.FilePath

	require.NoError(t, f.Read(resolver))
	assert.Equal(t, firstFilePath, f.FilePath)
	assert.Equal(t, filepath.Join(baseDir, "data.json"), f.FilePath)
}
