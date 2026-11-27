package pipeline

import (
	"testing"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/config"
	"github.com/updatecli/updatecli/pkg/core/pipeline/action"
	"github.com/updatecli/updatecli/pkg/core/pipeline/resource"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
	"github.com/updatecli/updatecli/pkg/core/pipeline/target"
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

func TestLocalScmDirectoryWrapsTheScmOfItsRepository(t *testing.T) {
	localDir := newOriginCheckout(t, "https://example.com/updatecli/updatecli.git")

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

// newOriginCheckout returns a git repository whose origin remote is origin.
func newOriginCheckout(t *testing.T, origin string) string {
	t.Helper()

	directory := t.TempDir()
	repository, err := git.PlainInit(directory, false)
	require.NoError(t, err)

	_, err = repository.CreateRemote(&gitconfig.RemoteConfig{Name: "origin", URLs: []string{origin}})
	require.NoError(t, err)

	return directory
}

func TestOnlyTheScmOfTheLocalCheckoutIsLocal(t *testing.T) {
	localDir := newOriginCheckout(t, "git@example.com:updatecli/updatecli.git")

	manifest := localScmManifest()
	manifest.Spec.SCMs["other"] = scm.Config{
		Kind: "git",
		Spec: map[string]interface{}{"url": "https://example.com/updatecli/website.git", "branch": "main"},
	}
	manifest.Spec.Targets = map[string]target.Config{
		"local": {ResourceConfig: resource.ResourceConfig{Kind: "shell", SCMID: "default"}},
		"other": {ResourceConfig: resource.ResourceConfig{Kind: "shell", SCMID: "other"}},
		"none":  {ResourceConfig: resource.ResourceConfig{Kind: "shell"}},
	}

	p := Pipeline{}
	require.NoError(t, p.Init(manifest, Options{LocalScmDirectory: localDir}))

	assert.IsType(t, &scm.Local{}, p.SCMs["default"].Handler, "the scm of the local checkout uses it")
	assert.NotEqual(t, localDir, p.SCMs["other"].Handler.GetDirectory(), "another repository is cloned")
	assert.False(t, scm.IsLocal(p.SCMs["other"].Handler))

	assert.False(t, p.targetOptions("local", p.Targets["local"]).DryRun, "the local checkout is changed")
	assert.True(t, p.targetOptions("other", p.Targets["other"]).DryRun, "a cloned repository is never changed")
	assert.False(t, p.targetOptions("none", p.Targets["none"]).DryRun, "a target without scm is unaffected")

	require.NoError(t, p.Update(), "Update rebuilds the scms and must match them the same way")

	assert.IsType(t, &scm.Local{}, p.SCMs["default"].Handler)
	assert.False(t, scm.IsLocal(p.SCMs["other"].Handler))
	assert.True(t, p.targetOptions("other", p.Targets["other"]).DryRun)
}

func TestTargetOptionsWithoutLocalCheckout(t *testing.T) {
	manifest := localScmManifest()
	manifest.Spec.Actions = nil
	manifest.Spec.Targets = map[string]target.Config{
		"default": {ResourceConfig: resource.ResourceConfig{Kind: "shell", SCMID: "default"}},
	}

	p := Pipeline{}
	require.NoError(t, p.Init(manifest, Options{}))

	assert.False(t, p.targetOptions("default", p.Targets["default"]).DryRun, "outside dev mode, nothing changes")
}

func TestTargetOptionsOnlyApplyLocalFileKinds(t *testing.T) {
	localDir := newOriginCheckout(t, "https://example.com/updatecli/updatecli.git")

	manifest := localScmManifest()
	manifest.Spec.Actions = nil
	manifest.Spec.Targets = map[string]target.Config{
		"file":          {ResourceConfig: resource.ResourceConfig{Kind: "file", SCMID: "default"}},
		"fileNoScm":     {ResourceConfig: resource.ResourceConfig{Kind: "file"}},
		"release":       {ResourceConfig: resource.ResourceConfig{Kind: "githubrelease", SCMID: "default"}},
		"tagNoScm":      {ResourceConfig: resource.ResourceConfig{Kind: "gittag"}},
		"branchNoScm":   {ResourceConfig: resource.ResourceConfig{Kind: "gitbranch"}},
		"releaseNoScm":  {ResourceConfig: resource.ResourceConfig{Kind: "gitea/release"}},
		"shellLocalScm": {ResourceConfig: resource.ResourceConfig{Kind: "shell", SCMID: "default"}},
	}

	options := Options{LocalScmDirectory: localDir}
	p := Pipeline{}
	require.NoError(t, p.Init(manifest, options))

	for id, dryRun := range map[string]bool{
		"file":          false,
		"fileNoScm":     false,
		"shellLocalScm": false,
		"release":       true,
		"tagNoScm":      true,
		"branchNoScm":   true,
		"releaseNoScm":  true,
	} {
		assert.Equal(t, dryRun, p.targetOptions(id, p.Targets[id]).DryRun, id)
	}

	outside := Pipeline{}
	require.NoError(t, outside.Init(manifest, Options{}))
	assert.False(t, outside.targetOptions("tagNoScm", outside.Targets["tagNoScm"]).DryRun, "outside dev mode, nothing changes")
}
