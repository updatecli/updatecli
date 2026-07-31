package dasel

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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
			assert.Error(t, f.Read(workingDir))
		})
	}
}
