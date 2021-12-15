package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/updatecli/updatecli/pkg/core/pipeline/scm"
)

func TestShell_Condition(t *testing.T) {
	tests := []struct {
		name              string
		command           string
		source            string
		wantResult        bool
		wantErr           bool
		wantCommand       string
		mockCommandResult commandResult
	}{
		{
			name:        "Successful Condition",
			command:     "echo Hello",
			source:      "1.2.3",
			wantResult:  true,
			wantErr:     false,
			wantCommand: "echo Hello 1.2.3",
			mockCommandResult: commandResult{
				ExitCode: 0,
				Stdout:   "Hello",
			},
		},
		{
			name:        "Failed Condition",
			command:     "ls",
			source:      "1.2.3",
			wantResult:  false,
			wantErr:     false,
			wantCommand: "ls 1.2.3",
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
				spec: ShellSpec{
					Command: tt.command,
				},
			}

			gotResult, err := s.Condition(tt.source)

			if tt.wantErr {
				assert.Error(t, err)
				assert.False(t, gotResult)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantResult, gotResult)

			assert.Equal(t, tt.wantCommand, mock.GotCommand.Cmd)
		})
	}
}

func TestShell_ConditionFromSCM(t *testing.T) {
	tests := []struct {
		name          string
		command       string
		source        string
		scmDir        string
		wantResult    bool
		wantErr       bool
		wantCommand   string
		commandResult commandResult
	}{
		{
			name:        "Successful Condition in existing SCM",
			command:     "echo Hello",
			source:      "1.2.3",
			scmDir:      "/dummy/dir",
			wantResult:  true,
			wantErr:     false,
			wantCommand: "echo Hello 1.2.3",
			commandResult: commandResult{
				ExitCode: 0,
				Stdout:   "Hello",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mce := MockCommandExecutor{
				Result: tt.commandResult,
			}
			ms := scm.MockScm{
				WorkingDir: tt.scmDir,
			}
			s := Shell{
				executor: &mce,
				spec: ShellSpec{
					Command: tt.command,
				},
			}

			gotResult, err := s.ConditionFromSCM(tt.source, &ms)

			if tt.wantErr {
				assert.Error(t, err)
				assert.False(t, gotResult)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantResult, gotResult)

			assert.Equal(t, tt.wantCommand, mce.GotCommand.Cmd)
			assert.Equal(t, tt.scmDir, mce.GotCommand.Dir)
		})
	}
}
