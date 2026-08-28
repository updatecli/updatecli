package toolversions

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/updatecli/updatecli/pkg/plugins/utils"
)

// TestFileContent_ReadPathContainment is the regression test for
// GHSA-hj4x-hm4v-7wpw at the toolversions sink: a path escaping the working
// directory must be rejected before any file is accessed. ContentRetriever is
// intentionally left nil so the test fails loudly if containment ever stops
// short-circuiting the file access.
func TestFileContent_ReadPathContainment(t *testing.T) {
	workingDir := t.TempDir()

	tests := []struct {
		name     string
		filePath string
	}{
		{name: "absolute path", filePath: filepath.Join(t.TempDir(), ".tool-versions")},
		{name: "dot dot traversal", filePath: filepath.Join("..", "..", ".tool-versions")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := FileContent{
				FilePath: tt.filePath,
			}
			assert.Error(t, f.Read(utils.Resolver{BaseDir: workingDir, Boundary: workingDir}))
		})
	}
}
