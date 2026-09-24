package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/plugins/scms/github"
)

func TestResolveScmDirectoryKeepsRawSpec(t *testing.T) {
	rawSpec := map[string]interface{}{"directory": "work", "branch": "main"}
	manifestScm := scm.Config{Kind: "git", Spec: rawSpec}

	resolvedScm := manifestScm
	resolveScmDirectory(&resolvedScm, "updatecli.d")

	assert.Equal(t, filepath.Join("updatecli.d", "work"), resolvedScm.Spec.(map[string]interface{})["directory"])
	assert.Equal(t, "work", rawSpec["directory"], "the manifest spec must keep its raw value")
}

func TestPinScmDirectory(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	absDir := filepath.Join(t.TempDir(), "checkout")

	tests := []struct {
		name      string
		spec      map[string]interface{}
		baseDir   string
		expectDir interface{}
	}{
		{
			name:      "relative directory without a base directory resolves from the cwd",
			spec:      map[string]interface{}{"directory": "work"},
			expectDir: filepath.Join(cwd, "work"),
		},
		{
			name:      "relative directory resolves from the base directory",
			spec:      map[string]interface{}{"directory": "work"},
			baseDir:   "updatecli.d",
			expectDir: filepath.Join(cwd, "updatecli.d", "work"),
		},
		{
			name:      "absolute directory is left alone",
			spec:      map[string]interface{}{"directory": absDir},
			baseDir:   "updatecli.d",
			expectDir: absDir,
		},
		{
			name:    "empty directory is left alone",
			spec:    map[string]interface{}{"branch": "main"},
			baseDir: "updatecli.d",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rawDir := tt.spec["directory"]

			pinned := PinScmDirectory(scm.Config{Kind: "git", Spec: tt.spec}, tt.baseDir)
			assert.Equal(t, tt.expectDir, pinned.Spec.(map[string]interface{})["directory"])
			assert.Equal(t, rawDir, tt.spec["directory"], "the input spec must not change")

			// A pinned directory must not move when the receiving manifest resolves it again.
			resolveScmDirectory(&pinned, "somewhere/else")
			assert.Equal(t, tt.expectDir, pinned.Spec.(map[string]interface{})["directory"])
		})
	}
}

// TestUpdateKeepsResolvedScmDirectory checks that Update, which rebuilds every scm before a
// resource or an action runs, resolves a relative scm directory like Init does.
func TestUpdateKeepsResolvedScmDirectory(t *testing.T) {
	manifest := &config.Config{
		Spec: config.Spec{
			PipelineID: "pipeline-id",
			SCMs: map[string]scm.Config{
				"default": {
					Kind: "git",
					Spec: map[string]interface{}{
						"url":       "https://example.com/updatecli/updatecli.git",
						"branch":    "main",
						"directory": "checkout",
					},
				},
			},
		},
	}
	manifest.SetBaseDir("updatecli.d")

	p := Pipeline{ID: "pipeline-id", Config: manifest, SCMs: map[string]scm.Scm{}}

	require.NoError(t, p.Update())

	assert.Equal(t, filepath.Join("updatecli.d", "checkout"), p.SCMs["default"].Handler.GetDirectory())
	assert.Equal(t, "checkout", manifest.Spec.SCMs["default"].Spec.(map[string]interface{})["directory"],
		"the manifest spec must keep its raw value, so each rebuild resolves it once")
}

// typedGithubSpec is an scm spec as githubsearch leaves it: a typed struct, not a raw map.
func typedGithubSpec(directory string) scm.Config {
	return scm.Config{
		Kind: "github",
		Spec: github.Spec{
			Owner:      "updatecli",
			Repository: "updatecli",
			Branch:     "main",
			Token:      "token",
			Directory:  directory,
		},
	}
}

// TestTypedScmSpecResolvesOnceAcrossInitAndUpdate checks that a typed spec gets the same
// directory when the pipeline is set up as when Update rebuilds it, which it does after
// Config.Update has turned the spec into a map.
func TestTypedScmSpecResolvesOnceAcrossInitAndUpdate(t *testing.T) {
	for _, baseDir := range []string{"", "updatecli.d"} {
		t.Run("base directory "+baseDir, func(t *testing.T) {
			manifest := &config.Config{
				Spec: config.Spec{
					PipelineID: "pipeline-id",
					SCMs:       map[string]scm.Config{"default": typedGithubSpec("repos")},
				},
			}
			manifest.SetBaseDir(baseDir)

			p := Pipeline{ID: "pipeline-id", Config: manifest, SCMs: map[string]scm.Scm{}}

			initScm, err := newScm(manifest.Spec.SCMs["default"], baseDir, "pipeline-id")
			require.NoError(t, err)

			require.NoError(t, p.Update())

			expectedDir := filepath.Join(baseDir, "repos")
			assert.Equal(t, expectedDir, initScm.Handler.GetDirectory())
			assert.Equal(t, expectedDir, p.SCMs["default"].Handler.GetDirectory())
		})
	}
}

func TestPinScmDirectoryTypedSpec(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	pinned := PinScmDirectory(typedGithubSpec("repos"), "")
	assert.Equal(t, filepath.Join(cwd, "repos"), pinned.Spec.(map[string]interface{})["directory"])

	// A pinned directory must not move when the generated manifest resolves it again.
	resolveScmDirectory(&pinned, filepath.Join(cwd, "repos"))
	assert.Equal(t, filepath.Join(cwd, "repos"), pinned.Spec.(map[string]interface{})["directory"])
}

func TestResolveScmDirectoryKeepsTypedSpecWithoutDirectory(t *testing.T) {
	scmConfig := typedGithubSpec("")

	resolveScmDirectory(&scmConfig, "updatecli.d")

	assert.IsType(t, github.Spec{}, scmConfig.Spec, "a spec with nothing to resolve must be left as it is")
}
