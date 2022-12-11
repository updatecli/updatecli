package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShell_Target(t *testing.T) {
	tests := []struct {
		name              string
		command           string
		source            string
		dryrun            bool
		wantChanged       bool
		wantErr           bool
		wantCommandInMock string
		commandResult     commandResult
		commandEnv        []string
	}{
		{
			name:              "runs a target that does not change anything and no dryrun",
			command:           "do_not_change.sh",
			source:            "1.2.3",
			wantChanged:       false,
			wantErr:           false,
			wantCommandInMock: "do_not_change.sh 1.2.3",
			commandResult: commandResult{
				ExitCode: 0,
				Stdout:   "",
			},
			commandEnv: []string{"DRY_RUN=false"},
		},
		{
			name:              "runs a target that changes a value and no dryrun",
			command:           "change.sh",
			source:            "1.2.3",
			wantChanged:       true,
			wantErr:           false,
			wantCommandInMock: "change.sh 1.2.3",
			commandResult: commandResult{
				ExitCode: 0,
				Stdout:   "1.2.3",
			},
			commandEnv: []string{"DRY_RUN=false"},
		},
		{
			name:              "runs a target with exit code 2 and no dryrun",
			command:           "change.sh",
			source:            "1.2.3",
			wantChanged:       false,
			wantErr:           true,
			wantCommandInMock: "change.sh 1.2.3",
			commandResult: commandResult{
				ExitCode: 2,
				Stderr:   "Error: unable to change value to 1.2.3.",
			},
			commandEnv: []string{"DRY_RUN=false"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := MockCommandExecutor{
				Result: tt.commandResult,
			}
			s := Shell{
				executor: &mock,
				spec: Spec{
					Command: tt.command,
				},
			}

			gotChanged, err := s.Target(tt.source, tt.dryrun)

			if tt.wantErr {
				assert.Error(t, err)
				assert.False(t, gotChanged)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantChanged, gotChanged)

			assert.Equal(t, tt.wantCommandInMock, mock.GotCommand.Cmd)
			for _, wantEnv := range tt.commandEnv {
				assert.Contains(t, mock.GotCommand.Env, wantEnv)
			}
		})
	}
}
