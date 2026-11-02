package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnforceLocalScm(t *testing.T) {
	local := Engine{}
	local.Options.Pipeline.LocalScmDirectory = "/work/repo"
	local.Options.Pipeline.Target.Commit = true
	local.Options.Pipeline.Target.Push = true

	local.enforceLocalScm()

	assert.True(t, local.Options.DisableUdashReport, "a local run never publishes to Udash")
	assert.False(t, local.Options.Pipeline.Target.Commit)
	assert.False(t, local.Options.Pipeline.Target.Push)

	remote := Engine{}
	remote.Options.Pipeline.Target.Commit = true
	remote.Options.Pipeline.Target.Push = true

	remote.enforceLocalScm()

	assert.False(t, remote.Options.DisableUdashReport, "a cloned checkout keeps publishing")
	assert.True(t, remote.Options.Pipeline.Target.Commit)
	assert.True(t, remote.Options.Pipeline.Target.Push)
}
