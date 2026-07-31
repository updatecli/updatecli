package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/result"
)

// TestFile_TargetPathContainment is the regression test for GHSA-hj4x-hm4v-7wpw.
// When a target path is (attacker) controlled — e.g. templated from a source
// output — it must never escape the SCM working directory, neither through an
// absolute path nor through ".." traversal. The pipeline must fail and nothing
// may be written outside the checkout.
func TestFile_TargetPathContainment(t *testing.T) {
	tests := []struct {
		name       string
		targetPath func(workingDir string) string
	}{
		{
			name: "absolute path escape is rejected",
			targetPath: func(_ string) string {
				return filepath.Join(t.TempDir(), "pwned_absolute.txt")
			},
		},
		{
			name: "dot dot traversal escape is rejected",
			targetPath: func(_ string) string {
				return filepath.Join("..", "..", "pwned_traversal.txt")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workingDir := t.TempDir()
			targetPath := tt.targetPath(workingDir)

			f, err := New(Spec{
				File:        targetPath,
				Content:     "PWNED",
				ForceCreate: true,
			})
			require.NoError(t, err)

			mockSCM := &scm.MockScm{WorkingDir: workingDir}

			gotErr := f.Target(context.Background(), "", mockSCM, false, &result.Target{})
			assert.Error(t, gotErr, "a path escaping the working directory must be rejected")

			// Whatever location the attack aimed at, the payload must not exist.
			escapePath := targetPath
			if !filepath.IsAbs(escapePath) {
				escapePath = filepath.Clean(filepath.Join(workingDir, targetPath))
			}
			_, statErr := os.Stat(escapePath)
			assert.Truef(t, os.IsNotExist(statErr),
				"payload must not be written outside the working directory: %q", escapePath)
		})
	}
}

// TestFile_TargetTemplatePathContainment covers the arbitrary-read half of
// GHSA-hj4x-hm4v-7wpw: a source templated spec.template must not be able to
// read files outside the working directory.
func TestFile_TargetTemplatePathContainment(t *testing.T) {
	workingDir := t.TempDir()

	// A secret sitting outside the checkout that the template read would try
	// to exfiltrate into the target file.
	secretDir := t.TempDir()
	secretPath := filepath.Join(secretDir, "secret.txt")
	require.NoError(t, os.WriteFile(secretPath, []byte("TOP SECRET"), 0o600))

	f, err := New(Spec{
		File:        "output.txt",
		Template:    secretPath, // absolute path outside the working directory
		ForceCreate: true,
	})
	require.NoError(t, err)

	mockSCM := &scm.MockScm{WorkingDir: workingDir}

	gotErr := f.Target(context.Background(), "", mockSCM, false, &result.Target{})
	assert.Error(t, gotErr, "reading a template outside the working directory must be rejected")
}
