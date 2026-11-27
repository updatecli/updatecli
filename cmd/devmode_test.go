package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/cmdoptions"
)

func TestAddDevFlagRegistration(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	addDevFlag(cmd)

	flag := cmd.Flags().Lookup("dev")
	require.NotNil(t, flag)
	assert.Equal(t, "false", flag.DefValue)
	assert.True(t, strings.HasPrefix(flag.Usage, "[Experimental"), "the help marks the flag experimental")
	assert.Contains(t, flag.Usage, devModePolicy)
}

func TestApplyDevFlag(t *testing.T) {
	previousDevMode, previousExperimental := devMode, cmdoptions.Experimental
	previousDirectory := e.Options.Pipeline.LocalScmDirectory
	t.Cleanup(func() {
		devMode, cmdoptions.Experimental = previousDevMode, previousExperimental
		e.Options.Pipeline.LocalScmDirectory = previousDirectory
	})

	e.Options.Pipeline.LocalScmDirectory = ""

	devMode, cmdoptions.Experimental = false, false
	require.NoError(t, applyDevFlag())
	assert.Empty(t, e.Options.Pipeline.LocalScmDirectory, "without --dev, scms are cloned")

	devMode, cmdoptions.Experimental = true, false
	require.ErrorIs(t, applyDevFlag(), errDevModeNeedsExperimental)
	assert.Empty(t, e.Options.Pipeline.LocalScmDirectory, "--dev does nothing without --experimental")

	repository := t.TempDir()
	_, err := git.PlainInit(repository, false)
	require.NoError(t, err)
	subDir := filepath.Join(repository, "updatecli.d")
	require.NoError(t, os.Mkdir(subDir, 0o750))
	t.Chdir(subDir)

	devMode, cmdoptions.Experimental = true, true
	require.NoError(t, applyDevFlag())
	assert.Equal(t, repository, e.Options.Pipeline.LocalScmDirectory, "the root of the repository is used from any subfolder")
}
