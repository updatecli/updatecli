package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func localScmManifest() *config.Config {
	return &config.Config{
		Spec: config.Spec{
			PipelineID: "pipeline-id",
			SCMs: map[string]scm.Config{
				"default": {
					Kind: "git",
					Spec: map[string]interface{}{
						"url":    "https://example.com/updatecli/updatecli.git",
						"branch": "main",
					},
				},
			},
			// A git scm can't back a pull request, so building this action would fail:
			// running against a local checkout must never build it.
			Actions: map[string]action.Config{
				"default": {Kind: "github/pullrequest", ScmID: "default"},
			},
		},
	}
}

func TestLocalScmDirectoryWrapsEveryScm(t *testing.T) {
	localDir := t.TempDir()

	p := Pipeline{}
	require.NoError(t, p.Init(localScmManifest(), Options{LocalScmDirectory: localDir}))

	assert.IsType(t, &scm.Local{}, p.SCMs["default"].Handler)
	assert.Equal(t, localDir, p.SCMs["default"].Handler.GetDirectory())
	assert.Empty(t, p.Actions, "actions are skipped against a local checkout")

	require.NoError(t, p.Update(), "Update rebuilds the scms and must keep them local")

	assert.IsType(t, &scm.Local{}, p.SCMs["default"].Handler)
	assert.Equal(t, localDir, p.SCMs["default"].Handler.GetDirectory())
	assert.Empty(t, p.Actions)
}

func TestWithoutLocalScmDirectoryActionsAreBuilt(t *testing.T) {
	p := Pipeline{}
	err := p.Init(localScmManifest(), Options{})

	assert.ErrorContains(t, err, `scm of kind "git" is not compatible with action of kind "github/pullrequest"`,
		"the action of the fixture is only valid when it is skipped")
}

func TestEnforceLocalScm(t *testing.T) {
	options := Options{LocalScmDirectory: t.TempDir()}
	options.Target.Commit = true
	options.Target.Push = true

	options.EnforceLocalScm()
	assert.False(t, options.Target.Commit)
	assert.False(t, options.Target.Push)

	remote := Options{}
	remote.Target.Commit = true
	remote.Target.Push = true

	remote.EnforceLocalScm()
	assert.True(t, remote.Target.Commit, "a cloned checkout keeps its commit option")
	assert.True(t, remote.Target.Push, "a cloned checkout keeps its push option")
}

func TestInitEnforcesLocalScm(t *testing.T) {
	options := Options{LocalScmDirectory: t.TempDir()}
	options.Target.Commit = true
	options.Target.Push = true

	p := Pipeline{}
	require.NoError(t, p.Init(localScmManifest(), options))

	assert.False(t, p.Options.Target.Commit)
	assert.False(t, p.Options.Target.Push)
}
