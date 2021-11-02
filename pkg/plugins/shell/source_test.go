package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShell_Source(t *testing.T) {
	tests := []struct {
		name          string
		command       string
		workingDir    string
		wantSource    string
		wantCommand   string
		wantErr       bool
		commandResult commandResult
	}{
		{
			name:        "Get a source from a successful command in working directory",
			command:     "echo Hello",
			workingDir:  "/home/ucli",
			wantSource:  "Hello",
			wantCommand: "echo Hello",
			wantErr:     false,
			commandResult: commandResult{
				ExitCode: 0,
				Stdout:   "Hello",
			},
		},
		{
			name:       "Raise an error with a failing command in working directory",
			command:    "false",
			workingDir: "/home/ucli",
			wantSource: "",
			wantErr:    true,
			commandResult: commandResult{
				ExitCode: 1,
			},
		},
		{
			name:       "Raise an error with an empty command in working directory",
			command:    "",
			workingDir: "/home/ucli",
			wantSource: "",
			wantErr:    true,
			commandResult: commandResult{
				ExitCode: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := MockCommandExecutor{
				Result: tt.commandResult,
			}
			s := Shell{
				executor: &mock,
				spec: ShellSpec{
					Command: tt.command,
				},
			}

			source, err := s.Source(tt.workingDir)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.wantSource, source)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantSource, source)

			assert.Equal(t, tt.wantCommand, mock.GotCommand.Cmd)
		})
	}
}
