package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNativeCommandExecutor_ExecuteCommand(t *testing.T) {
	var sut nativeCommandExecutor
	tests := []struct {
		name         string
		cmd          command
		wantExitCode int
		wantStdout   string
		wandStderr   string
		wantErr      bool
	}{
		{
			name: "Runs command with exit code 0",
			cmd: command{
				Cmd: "echo Hello",
			},
			wantExitCode: 0,
			wantStdout:   "Hello",
		},
		{
			name: "Runs command with exit code 1",
			cmd: command{
				Cmd: "false",
			},
			wantExitCode: 1,
		},
		{
			name: "Runs command with exit code 0 in a custom directory",
			cmd: command{
				Cmd: "pwd",
				// This directory should exist as we do not mock here. Avoid /tmp as it can be a link to another location
				Dir: "/",
			},
			wantExitCode: 0,
			wantStdout:   "/",
		},
		{
			name: "Runs command with exit code 0 in a nonexistent directory",
			cmd: command{
				Cmd: "pwd",
				Dir: "/toto",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := sut.ExecuteCommand(tt.cmd)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tt.wantExitCode, got.ExitCode)
			assert.Equal(t, tt.wantStdout, got.Stdout)
			assert.Equal(t, tt.wandStderr, got.Stderr)
		})
	}
}
