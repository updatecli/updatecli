package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	bashShell = "/bin/bash"
)

func TestShell_Condition(t *testing.T) {
	tests := []struct {
		name              string
		command           string
		shell             string
		source            string
		wantResult        bool
		wantErr           bool
		wantCommand       string
		mockCommandResult commandResult
	}{
		{
			name:        "Successful Condition",
			command:     "echo Hello",
			shell:       bashShell,
			source:      "1.2.3",
			wantResult:  true,
			wantErr:     false,
			wantCommand: bashShell + " " + wantedScriptFilename(t, "echo Hello 1.2.3"),
			mockCommandResult: commandResult{
				ExitCode: 0,
				Stdout:   "Hello",
			},
		},
		{
			name:        "Failed Condition",
			command:     "ls",
			shell:       bashShell,
			source:      "1.2.3",
			wantResult:  false,
			wantErr:     false,
			wantCommand: bashShell + " " + wantedScriptFilename(t, "ls 1.2.3"),
			mockCommandResult: commandResult{
				ExitCode: 1,
				Stderr:   "ls: 1.2.3: No such file or directory",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := MockCommandExecutor{
				Result: tt.mockCommandResult,
			}
			s := Shell{
				executor: &mock,
				spec: Spec{
					Command: tt.command,
					Shell:   tt.shell,
				},
				interpreter: tt.shell,
			}

			// InitSuccess
			gotErr := s.InitChangedIf()
			require.NoError(t, gotErr)

			gotResult, _, gotErr := s.Condition(tt.source, nil)

			if tt.wantErr {
				assert.Error(t, gotErr)
				return
			}

			require.NoError(t, gotErr)
			assert.Equal(t, tt.wantResult, gotResult)
			assert.Equal(t, tt.wantCommand, mock.GotCommand.Cmd)
		})
	}
}
